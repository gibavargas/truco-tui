package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
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

const (
	memberTTL            = 20 * time.Minute
	sessionTTL           = 6 * time.Hour
	joinTicketTTL        = 5 * time.Minute
	rateWindow           = 10 * time.Second
	rateCreatePerIP      = 20
	rateJoinPerIP        = 80
	rateMintPerSession   = 120
	rateRegisterPerSess  = 120
	rateTunnelPerSess    = 240
	maxSessions          = 4096
	maxMembersPerSession = 8
	maxTunnelsPerSession = 64
	gcInterval           = 1 * time.Minute
)

type relayServer struct {
	mu             sync.Mutex
	sessions       map[string]*relaySession
	tickets        map[string]*joinTicket
	authorityConns map[string]quic.Connection
	activeTunnels  map[string]int
	rateByIP       map[string]rateState
	rateBySession  map[string]rateState
	quicAddr       string
	tcpAddr        string
	ticketSecret   []byte
	metrics        relayMetrics

	tcpMu      sync.Mutex
	tcpSignals map[string]net.Conn
	tcpWaiting map[string]chan net.Conn
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
	authFailures       atomic.Int64
	ticketsMinted      atomic.Int64
	ticketsUsed        atomic.Int64
	ticketsExpired     atomic.Int64
	ticketsReplay      atomic.Int64
	rateLimited        atomic.Int64
	gcSessionsDeleted  atomic.Int64
	gcMembersDeleted   atomic.Int64
}

type relaySession struct {
	ID              string
	AdminToken      string
	NumPlayers      int
	Epoch           int
	AuthorityPeerID string
	Members         map[string]*memberState
	CreatedAt       time.Time
	ExpiresAt       time.Time
}

type memberState struct {
	PeerID     string
	Credential string
	Identity   string
	ExpiresAt  time.Time
	LastBeatAt time.Time
}

type joinTicket struct {
	TicketID      string
	SessionID     string
	DesiredRole   string
	TargetSeat    int
	PlayerSession string
	ExpiresAt     time.Time
	Used          bool
}

type rateState struct {
	WindowStart time.Time
	Count       int
}

type signedJoinTicket struct {
	TicketID      string `json:"ticket_id"`
	SessionID     string `json:"session_id"`
	DesiredRole   string `json:"desired_role,omitempty"`
	TargetSeat    int    `json:"target_seat,omitempty"`
	PlayerSession string `json:"player_session,omitempty"`
	ExpiresAtUnix int64  `json:"expires_at_unix"`
}

func newRelayServer(quicAddr string) *relayServer {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		panic("entropy source failed")
	}
	return &relayServer{
		sessions:       map[string]*relaySession{},
		tickets:        map[string]*joinTicket{},
		authorityConns: map[string]quic.Connection{},
		activeTunnels:  map[string]int{},
		rateByIP:       map[string]rateState{},
		rateBySession:  map[string]rateState{},
		quicAddr:       quicAddr,
		ticketSecret:   secret,
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
	ip := remoteIP(r.RemoteAddr)
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()
	if !allowRateLocked(s.rateByIP, "create:"+ip, rateCreatePerIP, rateWindow, now) {
		s.metrics.rateLimited.Add(1)
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "rate_limited"})
		return
	}
	if len(s.sessions) >= maxSessions {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "session_capacity_reached"})
		return
	}

	sessionID := randomHex(12)
	hostAdminToken := randomHex(24)
	hostPeerID := "seat-0"
	hostCredential := randomHex(24)
	sess := &relaySession{
		ID:              sessionID,
		AdminToken:      hostAdminToken,
		NumPlayers:      req.NumPlayers,
		Epoch:           1,
		AuthorityPeerID: hostPeerID,
		CreatedAt:       now,
		ExpiresAt:       now.Add(sessionTTL),
		Members: map[string]*memberState{
			hostPeerID: {
				PeerID:     hostPeerID,
				Credential: hostCredential,
				Identity:   strings.TrimSpace(req.HostIdentity),
				ExpiresAt:  now.Add(memberTTL),
				LastBeatAt: now,
			},
		},
	}
	s.sessions[sessionID] = sess
	s.metrics.sessionsCreated.Add(1)
	log.Printf("event=create_session req_id=%s session_id=%s players=%d", requestID(r), sessionID, req.NumPlayers)
	writeJSON(w, http.StatusOK, netrelay.CreateSessionResponse{
		SessionID:          sessionID,
		HostAdminToken:     hostAdminToken,
		HostPeerID:         hostPeerID,
		HostPeerCredential: hostCredential,
		AuthorityPeerID:    hostPeerID,
		Epoch:              1,
		QuicAddr:           s.quicAddr,
		TCPAddr:            s.tcpAddr,
		ExpiresAt:          sess.ExpiresAt,
	})
}

func (s *relayServer) handleMintJoinTicket(w http.ResponseWriter, r *http.Request) {
	var req netrelay.MintJoinTicketRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[req.SessionID]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "session_expired"})
		return
	}
	if !allowRateLocked(s.rateBySession, "mint:"+req.SessionID, rateMintPerSession, rateWindow, now) {
		s.metrics.rateLimited.Add(1)
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "rate_limited"})
		return
	}
	if subtle.ConstantTimeCompare([]byte(req.HostAdminToken), []byte(sess.AdminToken)) != 1 {
		s.metrics.authFailures.Add(1)
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "auth_failed"})
		return
	}
	if sess.ExpiresAt.Before(now) {
		writeJSON(w, http.StatusGone, map[string]any{"error": "session_expired"})
		return
	}
	tid := randomHex(18)
	t := &joinTicket{
		TicketID:      tid,
		SessionID:     req.SessionID,
		DesiredRole:   strings.TrimSpace(req.DesiredRole),
		TargetSeat:    req.TargetSeat,
		PlayerSession: strings.TrimSpace(req.PlayerSession),
		ExpiresAt:     now.Add(joinTicketTTL),
	}
	tok, err := s.signJoinTicket(t)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "ticket_sign_failed"})
		return
	}
	s.tickets[tid] = t
	s.metrics.ticketsMinted.Add(1)
	writeJSON(w, http.StatusOK, netrelay.MintJoinTicketResponse{JoinTicket: tok, ExpiresAt: t.ExpiresAt})
}

func (s *relayServer) handleJoinSession(w http.ResponseWriter, r *http.Request) {
	var req netrelay.JoinSessionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	now := time.Now()
	ip := remoteIP(r.RemoteAddr)

	s.mu.Lock()
	defer s.mu.Unlock()
	if !allowRateLocked(s.rateByIP, "join:"+ip, rateJoinPerIP, rateWindow, now) {
		s.metrics.rateLimited.Add(1)
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "rate_limited"})
		return
	}

	ticket, err := s.verifyJoinTicket(req.SessionID, req.JoinTicket, now)
	if err != nil {
		s.metrics.authFailures.Add(1)
		status := http.StatusUnauthorized
		msg := "auth_failed"
		switch err.Error() {
		case "ticket_expired":
			status = http.StatusGone
			msg = "ticket_expired"
		case "ticket_used":
			status = http.StatusConflict
			msg = "ticket_used"
		case "session_expired":
			status = http.StatusGone
			msg = "session_expired"
		}
		writeJSON(w, status, map[string]any{"error": msg})
		return
	}
	sess, ok := s.sessions[req.SessionID]
	if !ok || sess.ExpiresAt.Before(now) {
		writeJSON(w, http.StatusGone, map[string]any{"error": "session_expired"})
		return
	}
	if len(sess.Members) >= maxMembersPerSession {
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "session_full"})
		return
	}
	peerID := "peer-" + randomHex(10)
	if strings.TrimSpace(req.PlayerSession) != "" {
		peerID = "peer-" + strings.TrimSpace(req.PlayerSession)
	}
	cred := randomHex(24)
	exp := now.Add(memberTTL)
	sess.Members[peerID] = &memberState{
		PeerID:     peerID,
		Credential: cred,
		Identity:   strings.TrimSpace(req.PlayerName),
		ExpiresAt:  exp,
		LastBeatAt: now,
	}
	ticket.Used = true
	s.metrics.ticketsUsed.Add(1)
	s.metrics.joinsTotal.Add(1)
	writeJSON(w, http.StatusOK, netrelay.JoinSessionResponse{
		PeerID:           peerID,
		PeerCredential:   cred,
		AuthorityPeerID:  sess.AuthorityPeerID,
		Epoch:            sess.Epoch,
		QuicAddr:         s.quicAddr,
		TCPAddr:          s.tcpAddr,
		ExpiresAt:        exp,
		SessionExpiresAt: sess.ExpiresAt,
	})
}

func (s *relayServer) handlePublishAuthority(w http.ResponseWriter, r *http.Request) {
	var req netrelay.PublishAuthorityRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[req.SessionID]
	if !ok || sess.ExpiresAt.Before(now) {
		writeJSON(w, http.StatusGone, map[string]any{"error": "session_expired"})
		return
	}
	if subtle.ConstantTimeCompare([]byte(req.HostAdminToken), []byte(sess.AdminToken)) != 1 {
		s.metrics.authFailures.Add(1)
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "auth_failed"})
		return
	}
	if req.Epoch <= sess.Epoch {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "stale_epoch"})
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
	mem.ExpiresAt = now.Add(memberTTL)
	mem.LastBeatAt = now
	sess.AuthorityPeerID = req.HostPeerID
	sess.Epoch = req.Epoch
	delete(s.authorityConns, req.SessionID)
	s.metrics.authorityPublishes.Add(1)
	writeJSON(w, http.StatusOK, netrelay.PublishAuthorityResponse{
		AuthorityPeerID:    req.HostPeerID,
		Epoch:              req.Epoch,
		HostPeerCredential: cred,
		QuicAddr:           s.quicAddr,
		TCPAddr:            s.tcpAddr,
	})
}

func (s *relayServer) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var req netrelay.HeartbeatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[req.SessionID]
	if !ok || sess.ExpiresAt.Before(now) {
		writeJSON(w, http.StatusGone, map[string]any{"error": "session_expired"})
		return
	}
	mem, ok := sess.Members[req.PeerID]
	if !ok || mem.Credential != req.PeerCredential {
		s.metrics.authFailures.Add(1)
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "auth_failed"})
		return
	}
	mem.LastBeatAt = now
	mem.ExpiresAt = now.Add(memberTTL)
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
		"auth_failures":       s.metrics.authFailures.Load(),
		"tickets_minted":      s.metrics.ticketsMinted.Load(),
		"tickets_used":        s.metrics.ticketsUsed.Load(),
		"tickets_expired":     s.metrics.ticketsExpired.Load(),
		"tickets_replay":      s.metrics.ticketsReplay.Load(),
		"rate_limited":        s.metrics.rateLimited.Load(),
		"gc_sessions_deleted": s.metrics.gcSessionsDeleted.Load(),
		"gc_members_deleted":  s.metrics.gcMembersDeleted.Load(),
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
	fmt.Fprintf(w, "truco_relay_auth_failures_total %d\n", s.metrics.authFailures.Load())
	fmt.Fprintf(w, "truco_relay_tickets_minted_total %d\n", s.metrics.ticketsMinted.Load())
	fmt.Fprintf(w, "truco_relay_tickets_used_total %d\n", s.metrics.ticketsUsed.Load())
	fmt.Fprintf(w, "truco_relay_tickets_expired_total %d\n", s.metrics.ticketsExpired.Load())
	fmt.Fprintf(w, "truco_relay_tickets_replay_total %d\n", s.metrics.ticketsReplay.Load())
	fmt.Fprintf(w, "truco_relay_rate_limited_total %d\n", s.metrics.rateLimited.Load())
	fmt.Fprintf(w, "truco_relay_gc_sessions_deleted_total %d\n", s.metrics.gcSessionsDeleted.Load())
	fmt.Fprintf(w, "truco_relay_gc_members_deleted_total %d\n", s.metrics.gcMembersDeleted.Load())
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
			s.metrics.authFailures.Add(1)
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

// handleTCPConn processa uma conexão TCP+TLS no data plane do relay.
// Cada conexão carrega um único hello e, exceto pelo sinal do host, é pura
// tubulação de bytes após a autenticação.
func (s *relayServer) handleTCPConn(conn net.Conn) {
	handedOff := false
	defer func() {
		if handedOff {
			return
		}
		s.tcpMu.Lock()
		for sid, c := range s.tcpSignals {
			if c == conn {
				delete(s.tcpSignals, sid)
			}
		}
		s.tcpMu.Unlock()
		_ = conn.Close()
	}()
	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return
	}
	_ = conn.SetReadDeadline(time.Time{})
	var h netrelayHeartbeatTunnelHello
	if err := json.Unmarshal(bytesTrimSpace(line), &h); err != nil {
		return
	}
	switch h.Type {
	case "host_register":
		if err := s.handleTCPHostRegister(conn, h); err != nil {
			s.metrics.authFailures.Add(1)
			return
		}
		// Mantém a conexão de sinal aberta até o host desconectar.
		_, _ = reader.ReadBytes('\n')
	case "peer_tunnel":
		if err := s.handlePeerTunnel(reader, conn, h); err != nil {
			return
		}
	case "tunnel_accept":
		s.handleTCPTunnelAccept(conn, h, &handedOff)
	default:
	}
}

type netrelayHeartbeatTunnelHello struct {
	Type         string `json:"type"`
	SessionID    string `json:"session_id"`
	PeerID       string `json:"peer_id"`
	Credential   string `json:"credential"`
	TargetPeerID string `json:"target_peer_id"`
	TunnelID     string `json:"tunnel_id,omitempty"`
}

func (s *relayServer) handleHostRegister(conn quic.Connection, stream quic.Stream, h netrelayHeartbeatTunnelHello) error {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if !allowRateLocked(s.rateBySession, "register:"+h.SessionID, rateRegisterPerSess, rateWindow, now) {
		s.metrics.rateLimited.Add(1)
		return errors.New("rate_limited")
	}
	sess, ok := s.sessions[h.SessionID]
	if !ok || sess.ExpiresAt.Before(now) {
		return errors.New("session expired")
	}
	mem, ok := sess.Members[h.PeerID]
	if !ok || mem.Credential != h.Credential || mem.ExpiresAt.Before(now) {
		return errors.New("auth_failed")
	}
	if h.PeerID != sess.AuthorityPeerID {
		return errors.New("not authority")
	}
	s.authorityConns[h.SessionID] = conn
	_, _ = stream.Write([]byte("{\"ok\":true}\n"))
	_ = stream.Close()
	return nil
}

// handleTCPHostRegister autentica a conexão de sinal TCP do host e a registra
// como canal de anúncios de túneis (tunnel_open).
func (s *relayServer) handleTCPHostRegister(conn net.Conn, h netrelayHeartbeatTunnelHello) error {
	now := time.Now()
	s.mu.Lock()
	if !allowRateLocked(s.rateBySession, "register:"+h.SessionID, rateRegisterPerSess, rateWindow, now) {
		s.metrics.rateLimited.Add(1)
		s.mu.Unlock()
		return errors.New("rate_limited")
	}
	sess, ok := s.sessions[h.SessionID]
	if !ok || sess.ExpiresAt.Before(now) {
		s.mu.Unlock()
		return errors.New("session expired")
	}
	mem, ok := sess.Members[h.PeerID]
	if !ok || mem.Credential != h.Credential || mem.ExpiresAt.Before(now) {
		s.mu.Unlock()
		return errors.New("auth_failed")
	}
	if h.PeerID != sess.AuthorityPeerID {
		s.mu.Unlock()
		return errors.New("not authority")
	}
	s.mu.Unlock()

	ack, _ := json.Marshal(map[string]any{"ok": true})
	if _, err := conn.Write(append(ack, '\n')); err != nil {
		return err
	}
	s.tcpMu.Lock()
	s.tcpSignals[h.SessionID] = conn
	s.tcpMu.Unlock()
	return nil
}

// handleTCPTunnelAccept valida a conexão de dados discada pelo host TCP em
// resposta a um tunnel_open e a entrega ao par que aguardava.
func (s *relayServer) handleTCPTunnelAccept(conn net.Conn, h netrelayHeartbeatTunnelHello, handedOff *bool) {
	now := time.Now()
	s.mu.Lock()
	sess, ok := s.sessions[h.SessionID]
	if !ok || sess.ExpiresAt.Before(now) {
		s.mu.Unlock()
		return
	}
	mem, ok := sess.Members[h.PeerID]
	if !ok || mem.Credential != h.Credential || mem.ExpiresAt.Before(now) {
		s.metrics.authFailures.Add(1)
		s.mu.Unlock()
		return
	}
	if h.PeerID != sess.AuthorityPeerID {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	s.tcpMu.Lock()
	waiting, ok := s.tcpWaiting[h.TunnelID]
	delete(s.tcpWaiting, h.TunnelID)
	s.tcpMu.Unlock()
	if !ok {
		return
	}
	select {
	case waiting <- conn:
		*handedOff = true
	default:
		_ = conn.Close()
	}
}

func (s *relayServer) handlePeerTunnel(downstreamReader io.Reader, downstream io.Writer, h netrelayHeartbeatTunnelHello) error {
	now := time.Now()
	s.mu.Lock()
	if !allowRateLocked(s.rateBySession, "tunnel:"+h.SessionID, rateTunnelPerSess, rateWindow, now) {
		s.mu.Unlock()
		s.metrics.tunnelsFailed.Add(1)
		s.metrics.rateLimited.Add(1)
		return errors.New("rate_limited")
	}
	sess, ok := s.sessions[h.SessionID]
	if !ok || sess.ExpiresAt.Before(now) {
		s.mu.Unlock()
		s.metrics.tunnelsFailed.Add(1)
		return errors.New("session expired")
	}
	mem, ok := sess.Members[h.PeerID]
	if !ok || mem.Credential != h.Credential || mem.ExpiresAt.Before(now) {
		s.mu.Unlock()
		s.metrics.tunnelsFailed.Add(1)
		s.metrics.authFailures.Add(1)
		return errors.New("auth_failed")
	}
	targetPeer := strings.TrimSpace(h.TargetPeerID)
	if targetPeer == "" {
		targetPeer = sess.AuthorityPeerID
	}
	if targetPeer != sess.AuthorityPeerID {
		s.mu.Unlock()
		s.metrics.tunnelsFailed.Add(1)
		return errors.New("target peer is not current authority")
	}
	if s.activeTunnels[h.SessionID] >= maxTunnelsPerSession {
		s.mu.Unlock()
		s.metrics.tunnelsFailed.Add(1)
		s.metrics.rateLimited.Add(1)
		return errors.New("rate_limited")
	}
	s.activeTunnels[h.SessionID]++
	authorityConn, ok := s.authorityConns[h.SessionID]
	s.mu.Unlock()
	if !ok || authorityConn == nil {
		// Autoridade sem conexão QUIC: pode estar registrada via TCP.
		return s.handlePeerTunnelTCP(downstreamReader, downstream, h)
	}
	upstream, err := authorityConn.OpenStreamSync(context.Background())
	if err != nil {
		s.metrics.tunnelsFailed.Add(1)
		s.mu.Lock()
		s.activeTunnels[h.SessionID]--
		s.mu.Unlock()
		return err
	}
	s.metrics.tunnelsOpened.Add(1)
	s.bridgeTunnel(downstreamReader, downstream, upstream, h.SessionID)
	return nil
}

// handlePeerTunnelTCP resolve o lado da autoridade quando o host está
// registrado via TCP: anuncia tunnel_open na conexão de sinal e aguarda a
// conexão de dados (tunnel_accept) discada pelo host.
func (s *relayServer) handlePeerTunnelTCP(downstreamReader io.Reader, downstream io.Writer, h netrelayHeartbeatTunnelHello) error {
	s.tcpMu.Lock()
	signal, ok := s.tcpSignals[h.SessionID]
	if !ok {
		s.tcpMu.Unlock()
		s.metrics.tunnelsFailed.Add(1)
		s.mu.Lock()
		s.activeTunnels[h.SessionID]--
		s.mu.Unlock()
		return errors.New("authority unavailable")
	}
	tunnelID := randomHex(16)
	waiting := make(chan net.Conn, 1)
	s.tcpWaiting[tunnelID] = waiting
	s.tcpMu.Unlock()

	announce, _ := json.Marshal(map[string]any{
		"type":      "tunnel_open",
		"tunnel_id": tunnelID,
	})
	if _, err := signal.Write(append(announce, '\n')); err != nil {
		s.tcpMu.Lock()
		delete(s.tcpWaiting, tunnelID)
		s.tcpMu.Unlock()
		s.metrics.tunnelsFailed.Add(1)
		s.mu.Lock()
		s.activeTunnels[h.SessionID]--
		s.mu.Unlock()
		return err
	}
	select {
	case upstream := <-waiting:
		s.metrics.tunnelsOpened.Add(1)
		s.bridgeTunnel(downstreamReader, downstream, upstream, h.SessionID)
		return nil
	case <-time.After(10 * time.Second):
		s.tcpMu.Lock()
		delete(s.tcpWaiting, tunnelID)
		s.tcpMu.Unlock()
		s.metrics.tunnelsFailed.Add(1)
		s.mu.Lock()
		s.activeTunnels[h.SessionID]--
		s.mu.Unlock()
		return errors.New("authority dial timeout")
	}
}

// bridgeTunnel copia bytes nas duas direções e contabiliza o túnel.
func (s *relayServer) bridgeTunnel(downstreamReader io.Reader, downstream io.Writer, upstream io.ReadWriteCloser, sessionID string) {
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
	if closer, ok := downstream.(io.Closer); ok {
		_ = closer.Close()
	}
	s.mu.Lock()
	s.activeTunnels[sessionID]--
	s.mu.Unlock()
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

func (s *relayServer) gcLoop(ctx context.Context) {
	tk := time.NewTicker(gcInterval)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			s.gcOnce(time.Now())
		}
	}
}

func (s *relayServer) gcOnce(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for sid, sess := range s.sessions {
		activeMembers := 0
		for pid, mem := range sess.Members {
			if mem.ExpiresAt.Before(now) {
				delete(sess.Members, pid)
				s.metrics.gcMembersDeleted.Add(1)
				continue
			}
			activeMembers++
		}
		if sess.ExpiresAt.Before(now) || now.Sub(sess.CreatedAt) >= sessionTTL || activeMembers == 0 {
			delete(s.sessions, sid)
			delete(s.authorityConns, sid)
			delete(s.activeTunnels, sid)
			s.metrics.gcSessionsDeleted.Add(1)
		}
	}
	for tid, t := range s.tickets {
		if t.ExpiresAt.Before(now) {
			delete(s.tickets, tid)
			s.metrics.ticketsExpired.Add(1)
		}
	}
}

func (s *relayServer) signJoinTicket(t *joinTicket) (string, error) {
	claims := signedJoinTicket{
		TicketID:      t.TicketID,
		SessionID:     t.SessionID,
		DesiredRole:   t.DesiredRole,
		TargetSeat:    t.TargetSeat,
		PlayerSession: t.PlayerSession,
		ExpiresAtUnix: t.ExpiresAt.Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	h := hmac.New(sha256.New, s.ticketSecret)
	_, _ = h.Write(payload)
	sig := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (s *relayServer) verifyJoinTicket(sessionID, token string, now time.Time) (*joinTicket, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, errors.New("auth_failed")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("auth_failed")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("auth_failed")
	}
	h := hmac.New(sha256.New, s.ticketSecret)
	_, _ = h.Write(payload)
	if !hmac.Equal(sig, h.Sum(nil)) {
		return nil, errors.New("auth_failed")
	}
	var claims signedJoinTicket
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, errors.New("auth_failed")
	}
	if claims.SessionID != sessionID {
		return nil, errors.New("auth_failed")
	}
	t, ok := s.tickets[claims.TicketID]
	if !ok {
		return nil, errors.New("auth_failed")
	}
	if t.Used {
		s.metrics.ticketsReplay.Add(1)
		return nil, errors.New("ticket_used")
	}
	if now.After(t.ExpiresAt) || now.Unix() > claims.ExpiresAtUnix {
		s.metrics.ticketsExpired.Add(1)
		return nil, errors.New("ticket_expired")
	}
	sess, ok := s.sessions[sessionID]
	if !ok || sess.ExpiresAt.Before(now) {
		return nil, errors.New("session_expired")
	}
	return t, nil
}

func allowRateLocked(state map[string]rateState, key string, limit int, window time.Duration, now time.Time) bool {
	rs := state[key]
	if rs.WindowStart.IsZero() || now.Sub(rs.WindowStart) >= window {
		rs.WindowStart = now
		rs.Count = 0
	}
	rs.Count++
	state[key] = rs
	return rs.Count <= limit
}

func remoteIP(addr string) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil {
		return strings.TrimSpace(addr)
	}
	return strings.TrimSpace(host)
}

// defaultRelayTCPAddr deriva o endereço TCP do túnel a partir do QUIC:
// mesma porta quando o operador escolheu uma porta explícita, 9445 no padrão.
func defaultRelayTCPAddr(quicAddr string) string {
	host, port, err := net.SplitHostPort(strings.TrimSpace(quicAddr))
	if err != nil {
		return "127.0.0.1:9445"
	}
	if port == "9444" {
		return net.JoinHostPort(host, "9445")
	}
	return net.JoinHostPort(host, port)
}

func requestID(r *http.Request) string {
	v := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if v != "" {
		return v
	}
	return randomHex(6)
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("entropy source failed")
	}
	return hex.EncodeToString(b)
}

func bytesTrimSpace(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	if status >= 400 {
		if m, ok := v.(map[string]any); ok {
			if code, ok := m["error"].(string); ok {
				if _, exists := m["error_code"]; !exists {
					m["error_code"] = code
				}
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func relayTLSConfig() (*tls.Config, error) {
	certPath := strings.TrimSpace(os.Getenv("TRUCO_RELAY_TLS_CERT_FILE"))
	keyPath := strings.TrimSpace(os.Getenv("TRUCO_RELAY_TLS_KEY_FILE"))
	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("load managed cert: %w", err)
		}
		return &tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{netrelay.TunnelProto},
		}, nil
	}
	return relayTLSSelfSignedConfig()
}

func relayTLSSelfSignedConfig() (*tls.Config, error) {
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
			CommonName: "truco-relay-dev",
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
	cert := tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv}
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
	tcpAddr := os.Getenv("TRUCO_RELAY_TCP_ADDR")
	if strings.TrimSpace(tcpAddr) == "" {
		tcpAddr = defaultRelayTCPAddr(quicAddr)
	}

	server := newRelayServer(quicAddr)
	server.tcpAddr = tcpAddr
	server.tcpSignals = map[string]net.Conn{}
	server.tcpWaiting = map[string]chan net.Conn{}
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/create-session", server.handleCreateSession)
	mux.HandleFunc("/v2/mint-join-ticket", server.handleMintJoinTicket)
	mux.HandleFunc("/v2/join-session", server.handleJoinSession)
	mux.HandleFunc("/v2/publish-authority", server.handlePublishAuthority)
	mux.HandleFunc("/v2/heartbeat", server.handleHeartbeat)
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go server.gcLoop(ctx)
	go func() {
		for {
			conn, err := ql.Accept(context.Background())
			if err != nil {
				return
			}
			go server.handleQUICConn(conn)
		}
	}()

	tcpLn, err := tls.Listen("tcp", tcpAddr, tlsCfg)
	if err != nil {
		log.Fatalf("relay tcp listen: %v", err)
	}
	go func() {
		for {
			conn, err := tcpLn.Accept()
			if err != nil {
				return
			}
			go server.handleTCPConn(conn)
		}
	}()

	log.Printf("relay http=%s quic=%s tcp=%s", httpAddr, quicAddr, tcpAddr)
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
