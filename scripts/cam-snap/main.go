// cam-snap — microserviço de snapshot rápido da câmera IP.
//
// Endpoints:
//
//	GET /snap[?max_age=5s]  → JPEG fresco (usa cache se mais novo que max_age)
//	GET /last               → último frame em cache (instantâneo, qualquer idade)
//	GET /healthz            → status do serviço e da câmera
//
// Captura via HTTP (rápido ~300ms) com fallback RTSP+ffmpeg (~1s).
// Cache aquecido em background para /last quase instantâneo.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

type server struct {
	httpURL string
	rtspURL string
	user    string
	pass    string
	client  *http.Client
	ffmpeg  string

	capMu    sync.Mutex // serializa capturas
	cacheMu  sync.RWMutex
	lastData []byte
	lastTime time.Time
	lastSrc  string
}

func isJPEG(b []byte) bool {
	return len(b) > 3 && b[0] == 0xFF && b[1] == 0xD8 && b[len(b)-2] == 0xFF && b[len(b)-1] == 0xD9
}

func (s *server) captureHTTP(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.httpURL, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.user, s.pass)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, err
	}
	if !isJPEG(data) {
		return nil, errors.New("resposta não é JPEG")
	}
	return data, nil
}

func (s *server) captureRTSP(ctx context.Context) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, s.ffmpeg,
		"-rtsp_transport", "tcp", "-i", s.rtspURL,
		"-frames:v", "1", "-q:v", "2",
		"-f", "image2pipe", "pipe:1")
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		msg := errBuf.String()
		if len(msg) > 200 {
			msg = msg[len(msg)-200:]
		}
		return nil, fmt.Errorf("ffmpeg: %v: %s", err, msg)
	}
	data := out.Bytes()
	if !isJPEG(data) {
		return nil, errors.New("ffmpeg não produziu JPEG válido")
	}
	return data, nil
}

func (s *server) capture() (data []byte, src string, err error) {
	s.capMu.Lock()
	defer s.capMu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	if data, err = s.captureHTTP(ctx); err == nil {
		src = "http"
	} else {
		log.Printf("http falhou (%v), tentando rtsp", err)
		if data, err = s.captureRTSP(context.Background()); err == nil {
			src = "rtsp"
		} else {
			return nil, "", fmt.Errorf("http e rtsp falharam: %v", err)
		}
	}
	s.cacheMu.Lock()
	s.lastData, s.lastTime, s.lastSrc = data, time.Now(), src
	s.cacheMu.Unlock()
	return data, src, nil
}

func (s *server) cached(maxAge time.Duration) ([]byte, time.Time, string, bool) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	if s.lastData == nil {
		return nil, time.Time{}, "", false
	}
	// maxAge < 0: cache sempre válido (qualquer idade).
	if maxAge >= 0 && time.Since(s.lastTime) > maxAge {
		return nil, time.Time{}, "", false
	}
	return s.lastData, s.lastTime, s.lastSrc, true
}

func (s *server) serveSnap(w http.ResponseWriter, r *http.Request, maxAge time.Duration, skipCache bool) {
	if !skipCache {
		if data, ts, src, ok := s.cached(maxAge); ok {
			w.Header().Set("X-Source", src)
			w.Header().Set("X-Cache", "hit")
			w.Header().Set("X-Captured-At", ts.Format(time.RFC3339))
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(data)
			return
		}
	}
	data, src, err := s.capture()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("X-Source", src)
	w.Header().Set("X-Cache", "miss")
	w.Header().Set("X-Captured-At", time.Now().Format(time.RFC3339))
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(data)
}

func (s *server) healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	cam := "down"
	if _, err := s.captureHTTP(ctx); err == nil {
		cam = "up"
	}
	s.cacheMu.RLock()
	cached := len(s.lastData) > 0
	s.cacheMu.RUnlock()
	ageStr := ""
	if cached {
		ageStr = time.Since(s.lastTime).Round(time.Second).String()
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok":true,"camera":"%s","cached":%v,"cache_age":"%s"}`, cam, cached, ageStr)
}

func snapMaxAge(r *http.Request) time.Duration {
	q := r.URL.Query().Get("max_age")
	if q == "" {
		return 0 // estritamente fresco
	}
	if d, err := time.ParseDuration(q); err == nil {
		return d
	}
	if n, err := strconv.Atoi(q); err == nil {
		return time.Duration(n) * time.Second
	}
	return 0
}

func main() {
	var (
		listen  = flag.String("listen", "127.0.0.1:9378", "endereço de escuta")
		httpURL = flag.String("http-url", "http://192.168.100.64/image.jpg", "URL de snapshot HTTP")
		rtspURL = flag.String("rtsp-url", "rtsp://thingino:thingino@192.168.100.64:554/ch0", "URL RTSP (fallback)")
		user    = flag.String("user", "thingino", "usuário HTTP da câmera")
		pass    = flag.String("pass", "thingino", "senha HTTP da câmera")
		ffmpeg  = flag.String("ffmpeg", "/usr/bin/ffmpeg", "caminho do ffmpeg")
		warm    = flag.Duration("warm", 30*time.Second, "intervalo de aquecimento do cache (0 desliga)")
	)
	flag.Parse()

	s := &server{
		httpURL: *httpURL, rtspURL: *rtspURL, user: *user, pass: *pass, ffmpeg: *ffmpeg,
		client: &http.Client{Timeout: 5 * time.Second},
	}

	if *warm > 0 {
		go func() {
			for range time.Tick(*warm) {
				if _, _, err := s.capture(); err != nil {
					log.Printf("warm capture: %v", err)
				}
			}
		}()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/snap", func(w http.ResponseWriter, r *http.Request) { s.serveSnap(w, r, snapMaxAge(r), false) })
	mux.HandleFunc("/fresh", func(w http.ResponseWriter, r *http.Request) { s.serveSnap(w, r, 0, true) })
	mux.HandleFunc("/last", func(w http.ResponseWriter, r *http.Request) { s.serveSnap(w, r, -1, false) })
	mux.HandleFunc("/healthz", s.healthz)
	log.Printf("cam-snap escutando em %s (http=%s warm=%s)", *listen, *httpURL, *warm)
	if err := http.ListenAndServe(*listen, mux); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
