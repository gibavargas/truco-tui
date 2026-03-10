package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/quic-go/quic-go"

	"truco-tui/internal/netrelay"
)

type relayTestServer struct {
	httpURL  string
	quicAddr string
	httpSrv  *http.Server
	quicLn   *quic.Listener
}

func startRelayTestServer(t *testing.T) *relayTestServer {
	t.Helper()
	tlsCfg, err := relayTLSConfig()
	if err != nil {
		t.Fatalf("relayTLSConfig: %v", err)
	}
	ql, err := quic.ListenAddr("127.0.0.1:0", tlsCfg, &quic.Config{
		MaxIdleTimeout:       60 * time.Second,
		HandshakeIdleTimeout: 5 * time.Second,
		KeepAlivePeriod:      10 * time.Second,
	})
	if err != nil {
		t.Fatalf("quic.ListenAddr: %v", err)
	}
	server := newRelayServer(ql.Addr().String())
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/create-session", server.handleCreateSession)
	mux.HandleFunc("/v1/join-session", server.handleJoinSession)
	mux.HandleFunc("/v1/publish-authority", server.handlePublishAuthority)
	mux.HandleFunc("/v1/heartbeat", server.handleHeartbeat)
	mux.HandleFunc("/healthz", server.healthz)
	mux.HandleFunc("/metrics", server.metricsHandler)
	httpSrv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		TLSConfig:         tlsCfg,
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	go func() {
		for {
			conn, aerr := ql.Accept(context.Background())
			if aerr != nil {
				return
			}
			go server.handleQUICConn(conn)
		}
	}()
	go func() {
		_ = httpSrv.Serve(tls.NewListener(ln, tlsCfg))
	}()
	return &relayTestServer{
		httpURL:  "https://" + ln.Addr().String(),
		quicAddr: ql.Addr().String(),
		httpSrv:  httpSrv,
		quicLn:   ql,
	}
}

func (s *relayTestServer) Close(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = s.httpSrv.Shutdown(ctx)
	_ = s.quicLn.Close()
}

func insecureHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func TestRelayControlPlaneAndMetrics(t *testing.T) {
	srv := startRelayTestServer(t)
	defer srv.Close(t)

	created, err := netrelay.CreateSession(srv.httpURL, netrelay.CreateSessionRequest{
		HostIdentity: "Host",
		NumPlayers:   2,
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	joined, err := netrelay.JoinSession(srv.httpURL, netrelay.JoinSessionRequest{
		SessionID:    created.SessionID,
		SessionToken: created.SessionToken,
		PlayerName:   "Guest",
	})
	if err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	if err := netrelay.Heartbeat(srv.httpURL, netrelay.HeartbeatRequest{
		SessionID:      created.SessionID,
		PeerID:         joined.PeerID,
		PeerCredential: joined.PeerCredential,
		Epoch:          joined.Epoch,
	}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	_, err = netrelay.PublishAuthority(srv.httpURL, netrelay.PublishAuthorityRequest{
		SessionID:    created.SessionID,
		SessionToken: created.SessionToken,
		HostPeerID:   "seat-1",
		HostIdentity: "Guest",
		Epoch:        2,
	})
	if err != nil {
		t.Fatalf("PublishAuthority: %v", err)
	}

	client := insecureHTTPClient()
	res, err := client.Get(srv.httpURL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer res.Body.Close()
	var health map[string]any
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		t.Fatalf("decode healthz: %v", err)
	}
	if health["sessions_created"] == nil || health["joins_total"] == nil || health["authority_publishes"] == nil {
		t.Fatalf("missing metrics in healthz: %v", health)
	}

	mres, err := client.Get(srv.httpURL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer mres.Body.Close()
	body, _ := io.ReadAll(mres.Body)
	text := string(body)
	if !strings.Contains(text, "truco_relay_sessions_created") || !strings.Contains(text, "truco_relay_joins_total") {
		t.Fatalf("missing metric lines: %s", text)
	}
}

func TestRelayTunnelForwarding(t *testing.T) {
	srv := startRelayTestServer(t)
	defer srv.Close(t)

	created, err := netrelay.CreateSession(srv.httpURL, netrelay.CreateSessionRequest{
		HostIdentity: "Host",
		NumPlayers:   2,
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	hostAcceptor, err := netrelay.OpenHostAcceptor(context.Background(), created.QuicAddr, created.SessionID, created.HostPeerID, created.HostCredential)
	if err != nil {
		t.Fatalf("OpenHostAcceptor: %v", err)
	}
	defer hostAcceptor.Close()

	hostErr := make(chan error, 1)
	go func() {
		c, aerr := hostAcceptor.Accept()
		if aerr != nil {
			hostErr <- aerr
			return
		}
		defer c.Close()
		buf := make([]byte, 4)
		if _, rerr := io.ReadFull(c, buf); rerr != nil {
			hostErr <- rerr
			return
		}
		if string(buf) != "ping" {
			hostErr <- io.ErrUnexpectedEOF
			return
		}
		_, werr := c.Write([]byte("pong"))
		hostErr <- werr
	}()

	joined, err := netrelay.JoinSession(srv.httpURL, netrelay.JoinSessionRequest{
		SessionID:    created.SessionID,
		SessionToken: created.SessionToken,
		PlayerName:   "Guest",
	})
	if err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	peerConn, err := netrelay.OpenPeerTunnel(context.Background(), joined.QuicAddr, created.SessionID, joined.PeerID, joined.PeerCredential, joined.AuthorityPeerID)
	if err != nil {
		t.Fatalf("OpenPeerTunnel: %v", err)
	}
	defer peerConn.Close()

	if _, err := peerConn.Write([]byte("ping")); err != nil {
		t.Fatalf("peer write: %v", err)
	}
	resp := make([]byte, 4)
	if _, err := io.ReadFull(peerConn, resp); err != nil {
		t.Fatalf("peer read: %v", err)
	}
	if string(resp) != "pong" {
		t.Fatalf("response = %q, want pong", string(resp))
	}
	select {
	case err := <-hostErr:
		if err != nil {
			t.Fatalf("host stream error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting host stream")
	}
}
