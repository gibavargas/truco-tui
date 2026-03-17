package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
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
	sec      netrelay.ClientSecurity
}

func testTLSConfigForLocalhost(t *testing.T) (*tls.Config, *x509.CertPool, string) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatalf("rand.Int: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("CreateCertificate: %v", err)
	}
	pool := x509.NewCertPool()
	certObj, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("ParseCertificate: %v", err)
	}
	pool.AddCert(certObj)
	sum := sha256.Sum256(certObj.RawSubjectPublicKeyInfo)
	pin := base64.StdEncoding.EncodeToString(sum[:])
	return &tls.Config{
		MinVersion: tls.VersionTLS13,
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{der},
			PrivateKey:  priv,
		}},
		NextProtos: []string{netrelay.TunnelProto},
	}, pool, pin
}

func startRelayTestServer(t *testing.T) *relayTestServer {
	t.Helper()
	tlsCfg, pool, pin := testTLSConfigForLocalhost(t)
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
	mux.HandleFunc("/v2/create-session", server.handleCreateSession)
	mux.HandleFunc("/v2/mint-join-ticket", server.handleMintJoinTicket)
	mux.HandleFunc("/v2/join-session", server.handleJoinSession)
	mux.HandleFunc("/v2/publish-authority", server.handlePublishAuthority)
	mux.HandleFunc("/v2/heartbeat", server.handleHeartbeat)
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
		sec: netrelay.ClientSecurity{
			RootCAs:      pool,
			ServerName:   "127.0.0.1",
			RelaySPKIPin: pin,
		},
	}
}

func (s *relayTestServer) Close(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = s.httpSrv.Shutdown(ctx)
	_ = s.quicLn.Close()
}

func secureHTTPClient(sec netrelay.ClientSecurity) *http.Client {
	return &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    sec.RootCAs,
				ServerName: sec.ServerName,
			},
		},
	}
}

func TestRelayControlPlaneAndMetrics(t *testing.T) {
	srv := startRelayTestServer(t)
	defer srv.Close(t)

	created, err := netrelay.CreateSession(srv.httpURL, srv.sec, netrelay.CreateSessionRequest{
		HostIdentity: "Host",
		NumPlayers:   2,
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	ticket, err := netrelay.MintJoinTicket(srv.httpURL, srv.sec, netrelay.MintJoinTicketRequest{
		SessionID:      created.SessionID,
		HostAdminToken: created.HostAdminToken,
		PlayerName:     "Guest",
		DesiredRole:    "auto",
	})
	if err != nil {
		t.Fatalf("MintJoinTicket: %v", err)
	}
	joined, err := netrelay.JoinSession(srv.httpURL, srv.sec, netrelay.JoinSessionRequest{
		SessionID:   created.SessionID,
		JoinTicket:  ticket.JoinTicket,
		PlayerName:  "Guest",
		DesiredRole: "auto",
	})
	if err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	if err := netrelay.Heartbeat(srv.httpURL, srv.sec, netrelay.HeartbeatRequest{
		SessionID:      created.SessionID,
		PeerID:         joined.PeerID,
		PeerCredential: joined.PeerCredential,
		Epoch:          joined.Epoch,
	}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	_, err = netrelay.PublishAuthority(srv.httpURL, srv.sec, netrelay.PublishAuthorityRequest{
		SessionID:      created.SessionID,
		HostAdminToken: created.HostAdminToken,
		HostPeerID:     "seat-1",
		HostIdentity:   "Guest",
		Epoch:          2,
	})
	if err != nil {
		t.Fatalf("PublishAuthority: %v", err)
	}

	client := secureHTTPClient(srv.sec)
	res, err := client.Get(srv.httpURL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer res.Body.Close()
	var health map[string]any
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		t.Fatalf("decode healthz: %v", err)
	}
	if health["tickets_minted"] == nil || health["auth_failures"] == nil {
		t.Fatalf("missing metrics in healthz: %v", health)
	}

	mres, err := client.Get(srv.httpURL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer mres.Body.Close()
	body, _ := io.ReadAll(mres.Body)
	text := string(body)
	if !strings.Contains(text, "truco_relay_tickets_minted_total") || !strings.Contains(text, "truco_relay_auth_failures_total") {
		t.Fatalf("missing metric lines: %s", text)
	}
}

func TestRelayJoinTicketSingleUse(t *testing.T) {
	srv := startRelayTestServer(t)
	defer srv.Close(t)

	created, err := netrelay.CreateSession(srv.httpURL, srv.sec, netrelay.CreateSessionRequest{HostIdentity: "Host", NumPlayers: 2})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	ticket, err := netrelay.MintJoinTicket(srv.httpURL, srv.sec, netrelay.MintJoinTicketRequest{
		SessionID:      created.SessionID,
		HostAdminToken: created.HostAdminToken,
		PlayerName:     "Guest",
	})
	if err != nil {
		t.Fatalf("MintJoinTicket: %v", err)
	}
	_, err = netrelay.JoinSession(srv.httpURL, srv.sec, netrelay.JoinSessionRequest{SessionID: created.SessionID, JoinTicket: ticket.JoinTicket, PlayerName: "Guest"})
	if err != nil {
		t.Fatalf("JoinSession first: %v", err)
	}
	_, err = netrelay.JoinSession(srv.httpURL, srv.sec, netrelay.JoinSessionRequest{SessionID: created.SessionID, JoinTicket: ticket.JoinTicket, PlayerName: "Guest2"})
	if err == nil || !strings.Contains(err.Error(), "ticket_used") {
		t.Fatalf("JoinSession replay err=%v, want ticket_used", err)
	}
}

func TestRelaySPKIPinMismatchFails(t *testing.T) {
	srv := startRelayTestServer(t)
	defer srv.Close(t)
	badSec := srv.sec
	badSec.RelaySPKIPin = "deadbeef"
	_, err := netrelay.CreateSession(srv.httpURL, badSec, netrelay.CreateSessionRequest{HostIdentity: "Host", NumPlayers: 2})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "spki") {
		t.Fatalf("CreateSession err=%v, want SPKI mismatch", err)
	}
}

func TestRelayInvalidCertificateChainFails(t *testing.T) {
	srv := startRelayTestServer(t)
	defer srv.Close(t)
	badSec := srv.sec
	badSec.RootCAs = nil
	_, err := netrelay.CreateSession(srv.httpURL, badSec, netrelay.CreateSessionRequest{HostIdentity: "Host", NumPlayers: 2})
	if err == nil {
		t.Fatalf("CreateSession should fail with untrusted certificate")
	}
}

func TestRelayTicketSessionMismatchFails(t *testing.T) {
	srv := startRelayTestServer(t)
	defer srv.Close(t)
	a, err := netrelay.CreateSession(srv.httpURL, srv.sec, netrelay.CreateSessionRequest{HostIdentity: "A", NumPlayers: 2})
	if err != nil {
		t.Fatalf("CreateSession A: %v", err)
	}
	b, err := netrelay.CreateSession(srv.httpURL, srv.sec, netrelay.CreateSessionRequest{HostIdentity: "B", NumPlayers: 2})
	if err != nil {
		t.Fatalf("CreateSession B: %v", err)
	}
	ticket, err := netrelay.MintJoinTicket(srv.httpURL, srv.sec, netrelay.MintJoinTicketRequest{
		SessionID:      a.SessionID,
		HostAdminToken: a.HostAdminToken,
		PlayerName:     "Guest",
	})
	if err != nil {
		t.Fatalf("MintJoinTicket A: %v", err)
	}
	_, err = netrelay.JoinSession(srv.httpURL, srv.sec, netrelay.JoinSessionRequest{
		SessionID:  b.SessionID,
		JoinTicket: ticket.JoinTicket,
		PlayerName: "Guest",
	})
	if err == nil || !strings.Contains(err.Error(), "auth_failed") {
		t.Fatalf("JoinSession mismatch err=%v, want auth_failed", err)
	}
}

func TestRelayTunnelForwarding(t *testing.T) {
	srv := startRelayTestServer(t)
	defer srv.Close(t)

	created, err := netrelay.CreateSession(srv.httpURL, srv.sec, netrelay.CreateSessionRequest{
		HostIdentity: "Host",
		NumPlayers:   2,
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	hostAcceptor, err := netrelay.OpenHostAcceptor(context.Background(), srv.sec, created.QuicAddr, created.SessionID, created.HostPeerID, created.HostPeerCredential)
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

	ticket, err := netrelay.MintJoinTicket(srv.httpURL, srv.sec, netrelay.MintJoinTicketRequest{
		SessionID:      created.SessionID,
		HostAdminToken: created.HostAdminToken,
		PlayerName:     "Guest",
	})
	if err != nil {
		t.Fatalf("MintJoinTicket: %v", err)
	}
	joined, err := netrelay.JoinSession(srv.httpURL, srv.sec, netrelay.JoinSessionRequest{
		SessionID:  created.SessionID,
		JoinTicket: ticket.JoinTicket,
		PlayerName: "Guest",
	})
	if err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	peerConn, err := netrelay.OpenPeerTunnel(context.Background(), srv.sec, joined.QuicAddr, created.SessionID, joined.PeerID, joined.PeerCredential, joined.AuthorityPeerID)
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
