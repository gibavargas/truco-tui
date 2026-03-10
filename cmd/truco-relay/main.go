package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/quic-go/quic-go"

	"truco-tui/internal/netrelay"
)

type relayServer struct {
	mu             sync.Mutex
	sessions       map[string]*relaySession
	authorityConns map[string]quic.Connection
	quicAddr       string
	metrics        relayMetrics
}

type relayMetrics struct {
	sessionsCreated    atomic.Int64
	joinsTotal         atomic.Int64
	authorityPublishes atomic.Int64
	heartbeatsTotal    atomic.Int64
	tunnelsOpened      atomic.Int64
	tunnelsFailed      atomic.Int64
	tunnelBytesUp      atomic.Int64
	tunnelBytesDown    atomic.Int64
}

type relaySession struct {
	ID              string
	Token           string
	NumPlayers      int
	Epoch           int
	AuthorityPeerID string
	Members         map[string]*memberState
}

type memberState struct {
	PeerID     string
	Credential string
	Identity   string
	ExpiresAt  time.Time
	LastBeatAt time.Time
}

func newRelayServer(quicAddr string) *relayServer {
	return &relayServer{
		sessions:       map[string]*relaySession{},
		authorityConns: map[string]quic.Connection{},
		quicAddr:       quicAddr,
	}
}

func (s *relayServer) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req netrelay.CreateSessionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	if req.NumPlayers != 2 && req.NumPlayers != 4 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "num_players must be 2 or 4"})
		return
	}
	sessionID := randomHex(12)
	sessionToken := randomHex(24)
	hostPeerID := "seat-0"
	hostCredential := randomHex(24)
	now := time.Now()
	sess := &relaySession{
		ID:              sessionID,
		Token:           sessionToken,
		NumPlayers:      req.NumPlayers,
		Epoch:           1,
		AuthorityPeerID: hostPeerID,
		Members: map[string]*memberState{
			hostPeerID: {
				PeerID:     hostPeerID,
				Credential: hostCredential,
				Identity:   strings.TrimSpace(req.HostIdentity),
				ExpiresAt:  now.Add(20 * time.Minute),
				LastBeatAt: now,
			},
		},
	}
	s.mu.Lock()
	s.sessions[sessionID] = sess
	s.mu.Unlock()
	s.metrics.sessionsCreated.Add(1)
	log.Printf("event=create_session session_id=%s players=%d", sessionID, req.NumPlayers)
	writeJSON(w, http.StatusOK, netrelay.CreateSessionResponse{
		SessionID:       sessionID,
		SessionToken:    sessionToken,
		HostPeerID:      hostPeerID,
		HostCredential:  hostCredential,
		AuthorityPeerID: hostPeerID,
		Epoch:           1,
		QuicAddr:        s.quicAddr,
	})
}

func (s *relayServer) handleJoinSession(w http.ResponseWriter, r *http.Request) {
	var req netrelay.JoinSessionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[req.SessionID]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "session not found"})
		return
	}
	if req.SessionToken != sess.Token {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid session token"})
		return
	}
	peerID := "peer-" + randomHex(10)
	if strings.TrimSpace(req.PlayerSession) != "" {
		peerID = "peer-" + strings.TrimSpace(req.PlayerSession)
	}
	cred := randomHex(24)
	exp := time.Now().Add(20 * time.Minute)
	sess.Members[peerID] = &memberState{
		PeerID:     peerID,
		Credential: cred,
		Identity:   strings.TrimSpace(req.PlayerName),
		ExpiresAt:  exp,
		LastBeatAt: time.Now(),
	}
	s.metrics.joinsTotal.Add(1)
	log.Printf("event=join_session session_id=%s peer_id=%s authority=%s", req.SessionID, peerID, sess.AuthorityPeerID)
	writeJSON(w, http.StatusOK, netrelay.JoinSessionResponse{
		PeerID:          peerID,
		PeerCredential:  cred,
		AuthorityPeerID: sess.AuthorityPeerID,
		Epoch:           sess.Epoch,
		QuicAddr:        s.quicAddr,
		ExpiresAt:       exp,
	})
}

func (s *relayServer) handlePublishAuthority(w http.ResponseWriter, r *http.Request) {
	var req netrelay.PublishAuthorityRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[req.SessionID]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "session not found"})
		return
	}
	if req.SessionToken != sess.Token {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid session token"})
		return
	}
	if req.Epoch <= sess.Epoch {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "stale epoch"})
		return
	}
	if strings.TrimSpace(req.HostPeerID) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "host_peer_id required"})
		return
	}
	cred := randomHex(24)
	mem, ok := sess.Members[req.HostPeerID]
	if !ok {
		mem = &memberState{PeerID: req.HostPeerID}
		sess.Members[req.HostPeerID] = mem
	}
	mem.Credential = cred
	mem.Identity = strings.TrimSpace(req.HostIdentity)
	mem.ExpiresAt = time.Now().Add(20 * time.Minute)
	mem.LastBeatAt = time.Now()
	sess.AuthorityPeerID = req.HostPeerID
	sess.Epoch = req.Epoch
	delete(s.authorityConns, req.SessionID)
	s.metrics.authorityPublishes.Add(1)
	log.Printf("event=publish_authority session_id=%s host_peer_id=%s epoch=%d", req.SessionID, req.HostPeerID, req.Epoch)
	writeJSON(w, http.StatusOK, netrelay.PublishAuthorityResponse{
		AuthorityPeerID: req.HostPeerID,
		Epoch:           req.Epoch,
		HostCredential:  cred,
		QuicAddr:        s.quicAddr,
	})
}

func (s *relayServer) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var req netrelay.HeartbeatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[req.SessionID]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "session not found"})
		return
	}
	mem, ok := sess.Members[req.PeerID]
	if !ok || mem.Credential != req.PeerCredential {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid peer credential"})
		return
	}
	mem.LastBeatAt = time.Now()
	mem.ExpiresAt = time.Now().Add(20 * time.Minute)
	s.metrics.heartbeatsTotal.Add(1)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *relayServer) healthz(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":                  true,
		"sessions":            len(s.sessions),
		"active_authorities":  len(s.authorityConns),
		"sessions_created":    s.metrics.sessionsCreated.Load(),
		"joins_total":         s.metrics.joinsTotal.Load(),
		"authority_publishes": s.metrics.authorityPublishes.Load(),
		"heartbeats_total":    s.metrics.heartbeatsTotal.Load(),
		"tunnels_opened":      s.metrics.tunnelsOpened.Load(),
		"tunnels_failed":      s.metrics.tunnelsFailed.Load(),
		"tunnel_bytes_up":     s.metrics.tunnelBytesUp.Load(),
		"tunnel_bytes_down":   s.metrics.tunnelBytesDown.Load(),
	})
}

func (s *relayServer) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "truco_relay_sessions_created %d\n", s.metrics.sessionsCreated.Load())
	fmt.Fprintf(w, "truco_relay_joins_total %d\n", s.metrics.joinsTotal.Load())
	fmt.Fprintf(w, "truco_relay_authority_publishes_total %d\n", s.metrics.authorityPublishes.Load())
	fmt.Fprintf(w, "truco_relay_heartbeats_total %d\n", s.metrics.heartbeatsTotal.Load())
	fmt.Fprintf(w, "truco_relay_tunnels_opened_total %d\n", s.metrics.tunnelsOpened.Load())
	fmt.Fprintf(w, "truco_relay_tunnels_failed_total %d\n", s.metrics.tunnelsFailed.Load())
	fmt.Fprintf(w, "truco_relay_tunnel_bytes_up_total %d\n", s.metrics.tunnelBytesUp.Load())
	fmt.Fprintf(w, "truco_relay_tunnel_bytes_down_total %d\n", s.metrics.tunnelBytesDown.Load())
}

func (s *relayServer) handleQUICConn(conn quic.Connection) {
	defer s.cleanupConn(conn)
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			return
		}
		go s.handleQUICStream(conn, stream)
	}
}

func (s *relayServer) handleQUICStream(conn quic.Connection, stream quic.Stream) {
	reader := bufio.NewReader(stream)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		_ = stream.Close()
		return
	}
	var h netrelayHeartbeatTunnelHello
	if err := json.Unmarshal(bytesTrimSpace(line), &h); err != nil {
		_ = stream.Close()
		return
	}
	switch h.Type {
	case "host_register":
		if err := s.handleHostRegister(conn, stream, h); err != nil {
			_ = stream.Close()
			return
		}
		return
	case "peer_tunnel":
		if err := s.handlePeerTunnel(reader, stream, h); err != nil {
			_ = stream.Close()
			return
		}
		return
	default:
		_ = stream.Close()
		return
	}
}

type netrelayHeartbeatTunnelHello struct {
	Type         string `json:"type"`
	SessionID    string `json:"session_id"`
	PeerID       string `json:"peer_id"`
	Credential   string `json:"credential"`
	TargetPeerID string `json:"target_peer_id"`
}

func (s *relayServer) handleHostRegister(conn quic.Connection, stream quic.Stream, h netrelayHeartbeatTunnelHello) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[h.SessionID]
	if !ok {
		return errors.New("session not found")
	}
	mem, ok := sess.Members[h.PeerID]
	if !ok || mem.Credential != h.Credential || mem.ExpiresAt.Before(time.Now()) {
		return errors.New("invalid credential")
	}
	if h.PeerID != sess.AuthorityPeerID {
		return errors.New("not authority")
	}
	s.authorityConns[h.SessionID] = conn
	_, _ = stream.Write([]byte("{\"ok\":true}\n"))
	_ = stream.Close()
	return nil
}

func (s *relayServer) handlePeerTunnel(downstreamReader io.Reader, downstream quic.Stream, h netrelayHeartbeatTunnelHello) error {
	s.mu.Lock()
	sess, ok := s.sessions[h.SessionID]
	if !ok {
		s.mu.Unlock()
		return errors.New("session not found")
	}
	mem, ok := sess.Members[h.PeerID]
	if !ok || mem.Credential != h.Credential || mem.ExpiresAt.Before(time.Now()) {
		s.mu.Unlock()
		return errors.New("invalid credential")
	}
	targetPeer := strings.TrimSpace(h.TargetPeerID)
	if targetPeer == "" {
		targetPeer = sess.AuthorityPeerID
	}
	if targetPeer != sess.AuthorityPeerID {
		s.mu.Unlock()
		return errors.New("target peer is not current authority")
	}
	authorityConn, ok := s.authorityConns[h.SessionID]
	s.mu.Unlock()
	if !ok || authorityConn == nil {
		s.metrics.tunnelsFailed.Add(1)
		return errors.New("authority unavailable")
	}
	upstream, err := authorityConn.OpenStreamSync(context.Background())
	if err != nil {
		s.metrics.tunnelsFailed.Add(1)
		return err
	}
	s.metrics.tunnelsOpened.Add(1)
	log.Printf("event=open_tunnel session_id=%s from=%s to=%s", h.SessionID, h.PeerID, targetPeer)
	errc := make(chan error, 2)
	go func() {
		n, e := io.Copy(upstream, downstreamReader)
		s.metrics.tunnelBytesUp.Add(n)
		errc <- e
	}()
	go func() {
		n, e := io.Copy(downstream, upstream)
		s.metrics.tunnelBytesDown.Add(n)
		errc <- e
	}()
	<-errc
	_ = upstream.Close()
	_ = downstream.Close()
	return nil
}

func (s *relayServer) cleanupConn(conn quic.Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for sid, c := range s.authorityConns {
		if c == conn {
			delete(s.authorityConns, sid)
		}
	}
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "rnd-fallback"
	}
	return hex.EncodeToString(b)
}

func bytesTrimSpace(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func relayTLSConfig() (*tls.Config, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: "truco-relay",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}
	cert := tls.Certificate{
		Certificate: [][]byte{der},
		PrivateKey:  priv,
	}
	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{netrelay.TunnelProto},
	}, nil
}

func main() {
	httpAddr := os.Getenv("TRUCO_RELAY_HTTP_ADDR")
	if strings.TrimSpace(httpAddr) == "" {
		httpAddr = "127.0.0.1:9443"
	}
	quicAddr := os.Getenv("TRUCO_RELAY_QUIC_ADDR")
	if strings.TrimSpace(quicAddr) == "" {
		quicAddr = "127.0.0.1:9444"
	}

	server := newRelayServer(quicAddr)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/create-session", server.handleCreateSession)
	mux.HandleFunc("/v1/join-session", server.handleJoinSession)
	mux.HandleFunc("/v1/publish-authority", server.handlePublishAuthority)
	mux.HandleFunc("/v1/heartbeat", server.handleHeartbeat)
	mux.HandleFunc("/healthz", server.healthz)
	mux.HandleFunc("/metrics", server.metricsHandler)

	tlsCfg, err := relayTLSConfig()
	if err != nil {
		log.Fatalf("relay tls config: %v", err)
	}

	ql, err := quic.ListenAddr(quicAddr, tlsCfg, &quic.Config{
		MaxIdleTimeout:       60 * time.Second,
		HandshakeIdleTimeout: 5 * time.Second,
		KeepAlivePeriod:      10 * time.Second,
	})
	if err != nil {
		log.Fatalf("relay quic listen: %v", err)
	}
	go func() {
		for {
			conn, err := ql.Accept(context.Background())
			if err != nil {
				return
			}
			go server.handleQUICConn(conn)
		}
	}()

	log.Printf("relay http=%s quic=%s", httpAddr, quicAddr)
	httpSrv := &http.Server{
		Addr:              httpAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		TLSConfig:         tlsCfg,
	}
	ln, err := net.Listen("tcp", httpAddr)
	if err != nil {
		log.Fatalf("relay http listen: %v", err)
	}
	tlsListener := tls.NewListener(ln, tlsCfg)
	if err := httpSrv.Serve(tlsListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("relay http serve: %v", err)
	}
}
