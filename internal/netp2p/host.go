package netp2p

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"net"
	"strings"
	"sync"
	"time"

	"truco-tui/internal/truco"
)

// ClientAction representa uma ação de jogo enviada por um cliente remoto.
type ClientAction struct {
	Seat      int
	Action    string
	CardIndex int
}

// HostSession gerencia lobby, chat e transporte de ações/estado da partida.
type HostSession struct {
	mu              sync.Mutex
	ctx             context.Context
	cancel          context.CancelFunc
	ln              net.Listener
	cfg             HostConfig
	tlsNotAfter     time.Time
	tlsExpiryWarned bool
	token           string
	handoffPort     int
	inviteBase      InviteKey
	numPlayers      int
	hostName        string
	slots           []string
	clients         map[int]net.Conn
	seatID          map[int]string
	peerHosts       map[int]string
	tableHostSeat   int
	hostVotes       map[int]int
	replaceInvites  map[string]int
	acceptRate      map[string]acceptRateState
	seatRate        map[int]seatRateState
	lastPong        map[int]time.Time
	lastState       map[int]truco.Snapshot
	events          chan string
	actions         chan ClientAction
	closed          bool
	started         bool
}

type HostConfig struct {
	HeartbeatInterval   time.Duration
	HeartbeatTimeout    time.Duration
	ShutdownDrainWait   time.Duration
	TLSExpiryWarnBefore time.Duration
	HandoffPort         int
	AdvertiseHost       string
}

const (
	defaultHostHeartbeatInterval   = 5 * time.Second
	defaultHostHeartbeatTimeout    = 45 * time.Second
	defaultHostShutdownDrainWait   = 120 * time.Millisecond
	defaultHostTLSExpiryWarnBefore = 6 * time.Hour

	hostAcceptWindow         = 2 * time.Second
	hostAcceptMaxPerIPWindow = 20
	seatRateWindow           = 1 * time.Second
	seatMaxActionsPerWindow  = 12
	seatMaxChatPerWindow     = 6
	seatMaxControlPerWindow  = 4
)

type acceptRateState struct {
	windowStart time.Time
	count       int
}

type seatRateState struct {
	windowStart time.Time
	actions     int
	chat        int
	control     int
}

func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func NewHostSession(bindAddr, hostName string, numPlayers int) (*HostSession, string, error) {
	return NewHostSessionWithConfig(bindAddr, hostName, numPlayers, HostConfig{})
}

func NewHostSessionWithConfig(bindAddr, hostName string, numPlayers int, cfg HostConfig) (*HostSession, string, error) {
	token, err := randomToken()
	if err != nil {
		return nil, "", err
	}
	return newHostSession(bindAddr, hostName, numPlayers, token, cfg)
}

type RecoveredHostState struct {
	Token          string
	Slots          []string
	SeatSessionIDs map[int]string
	PeerHosts      map[int]string
	TableHostSeat  int
}

func NewRecoveredHostSession(bindAddr, hostName string, numPlayers int, state RecoveredHostState, cfg HostConfig) (*HostSession, string, error) {
	token := strings.TrimSpace(state.Token)
	if token == "" {
		return nil, "", errors.New("token de recuperação inválido")
	}
	hs, key, err := newHostSession(bindAddr, hostName, numPlayers, token, cfg)
	if err != nil {
		return nil, "", err
	}
	hs.mu.Lock()
	if len(state.Slots) != numPlayers {
		hs.mu.Unlock()
		_ = hs.Close()
		return nil, "", errors.New("slots de recuperação inválidos")
	}
	copy(hs.slots, state.Slots)
	if hs.slots[0] == "" {
		hs.slots[0] = hostName
	}
	hs.hostName = hs.slots[0]
	hs.started = true
	hs.tableHostSeat = 0
	if state.TableHostSeat >= 0 && state.TableHostSeat < numPlayers {
		hs.tableHostSeat = state.TableHostSeat
	}
	if hs.seatID == nil {
		hs.seatID = map[int]string{}
	}
	for seat := 0; seat < numPlayers; seat++ {
		if sid, ok := state.SeatSessionIDs[seat]; ok && strings.TrimSpace(sid) != "" {
			hs.seatID[seat] = sid
			continue
		}
		if seat == 0 {
			continue
		}
		sid, sidErr := randomToken()
		if sidErr != nil {
			hs.mu.Unlock()
			_ = hs.Close()
			return nil, "", sidErr
		}
		hs.seatID[seat] = sid
	}
	for seat, host := range state.PeerHosts {
		if seat < 0 || seat >= numPlayers {
			continue
		}
		if normalized := normalizeAdvertiseHost(host); normalized != "" {
			hs.peerHosts[seat] = normalized
		}
	}
	hs.mu.Unlock()
	return hs, key, nil
}

func newHostSession(bindAddr, hostName string, numPlayers int, token string, cfg HostConfig) (*HostSession, string, error) {
	if numPlayers != 2 && numPlayers != 4 {
		return nil, "", errors.New("numPlayers deve ser 2 ou 4")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, "", errors.New("token inválido")
	}
	cfg = cfg.normalized()
	if cfg.HandoffPort == 0 {
		cfg.HandoffPort = HandoffPortForToken(token)
	}
	ln, fingerprint, tlsNotAfter, err := generateTLSListener(bindAddr, token)
	if err != nil {
		return nil, "", err
	}
	ctx, cancel := context.WithCancel(context.Background())
	listenAddr := ln.Addr().String()
	inviteAddr := buildInviteAddr(listenAddr, cfg.AdvertiseHost)
	inviteBase := InviteKey{Addr: inviteAddr, Token: token, Fingerprint: fingerprint}
	hs := &HostSession{
		ctx:            ctx,
		cancel:         cancel,
		ln:             ln,
		cfg:            cfg,
		tlsNotAfter:    tlsNotAfter,
		token:          token,
		handoffPort:    cfg.HandoffPort,
		inviteBase:     inviteBase,
		numPlayers:     numPlayers,
		hostName:       hostName,
		slots:          make([]string, numPlayers),
		clients:        map[int]net.Conn{},
		seatID:         map[int]string{},
		peerHosts:      map[int]string{},
		tableHostSeat:  0,
		hostVotes:      map[int]int{},
		replaceInvites: map[string]int{},
		acceptRate:     map[string]acceptRateState{},
		seatRate:       map[int]seatRateState{},
		lastPong:       map[int]time.Time{},
		lastState:      map[int]truco.Snapshot{},
		events:         make(chan string, 128),
		actions:        make(chan ClientAction, 256),
	}
	hs.slots[0] = hostName
	if hostID, idErr := randomToken(); idErr == nil {
		hs.seatID[0] = hostID
	}
	if normalized := normalizeAdvertiseHost(cfg.AdvertiseHost); normalized != "" {
		hs.peerHosts[0] = normalized
	} else if host, _, splitErr := net.SplitHostPort(inviteAddr); splitErr == nil {
		if normalized = normalizeAdvertiseHost(host); normalized != "" {
			hs.peerHosts[0] = normalized
		}
	}

	key, err := EncodeInviteKey(inviteBase)
	if err != nil {
		if cerr := ln.Close(); cerr != nil {
			logNetf("close listener (key encode failure): %v", cerr)
		}
		return nil, "", err
	}
	go hs.acceptLoop()
	go hs.heartbeatLoop()
	return hs, key, nil
}

func HandoffPortForToken(token string) int {
	sum := crc32.ChecksumIEEE([]byte(token))
	return 39000 + int(sum%10000)
}

func (c HostConfig) normalized() HostConfig {
	if c.HeartbeatInterval <= 0 {
		c.HeartbeatInterval = defaultHostHeartbeatInterval
	}
	if c.HeartbeatTimeout <= 0 {
		c.HeartbeatTimeout = defaultHostHeartbeatTimeout
	}
	if c.ShutdownDrainWait <= 0 {
		c.ShutdownDrainWait = defaultHostShutdownDrainWait
	}
	if c.TLSExpiryWarnBefore <= 0 {
		c.TLSExpiryWarnBefore = defaultHostTLSExpiryWarnBefore
	}
	return c
}

func (h *HostSession) Events() <-chan string        { return h.events }
func (h *HostSession) Actions() <-chan ClientAction { return h.actions }

func (h *HostSession) Close() error {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return nil
	}
	h.closed = true
	h.cancel()
	for _, c := range h.clients {
		_ = writeMessage(c, Message{Type: "shutdown", Text: "Host encerrou a sessão."})
	}
	h.mu.Unlock()

	// Pequena janela para o pacote de shutdown sair do buffer do SO.
	time.Sleep(h.cfg.ShutdownDrainWait)

	h.mu.Lock()
	for _, c := range h.clients {
		closeConnWithLog(c, "host close")
	}
	if err := h.ln.Close(); err != nil {
		logNetf("close listener (host close): %v", err)
	}
	h.mu.Unlock()
	return nil
}

func (h *HostSession) Slots() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]string, len(h.slots))
	copy(out, h.slots)
	return out
}

func (h *HostSession) IsFull() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, s := range h.slots {
		if s == "" {
			return false
		}
	}
	return true
}

func (h *HostSession) StartGame() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.started {
		return errors.New("partida já iniciada")
	}
	for _, s := range h.slots {
		if s == "" {
			return errors.New("lobby ainda não está cheio")
		}
	}
	h.started = true
	h.broadcastLocked(Message{Type: "game_start", Slots: append([]string{}, h.slots...), NumPlayers: h.numPlayers})
	h.sendEventLocked("Partida iniciada.")
	return nil
}

func (h *HostSession) GameStarted() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.started
}

func (h *HostSession) CurrentHostSeat() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.tableHostSeat
}

func (h *HostSession) IsSeatConnected(seat int) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if seat == 0 {
		return true
	}
	_, ok := h.clients[seat]
	return ok
}

func (h *HostSession) ConnectedSeats() map[int]bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make(map[int]bool, len(h.clients)+1)
	out[0] = true
	for seat := range h.clients {
		out[seat] = true
	}
	return out
}

func (h *HostSession) CastHostVote(voterSeat, candidateSeat int) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.castHostVoteLocked(voterSeat, candidateSeat)
}

func (h *HostSession) RequestReplacementInvite(requesterSeat, targetSeat int) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.createReplacementInviteLocked(requesterSeat, targetSeat)
}

func (h *HostSession) chooseSlotLocked(role string) int {
	if h.numPlayers == 2 {
		if h.slots[1] == "" {
			return 1
		}
		return -1
	}
	partnerSlots := []int{2}
	opponentSlots := []int{1, 3}
	try := func(slots []int) int {
		for _, s := range slots {
			if s < len(h.slots) && h.slots[s] == "" {
				return s
			}
		}
		return -1
	}
	switch role {
	case "partner":
		if s := try(partnerSlots); s != -1 {
			return s
		}
		return try(opponentSlots)
	case "opponent":
		if s := try(opponentSlots); s != -1 {
			return s
		}
		return try(partnerSlots)
	default:
		for i := 1; i < len(h.slots); i++ {
			if h.slots[i] == "" {
				return i
			}
		}
		return -1
	}
}

func remoteIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

func normalizeAdvertiseHost(host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")
	if host == "" {
		return ""
	}
	if strings.Contains(host, ":") {
		parsed, _, err := net.SplitHostPort(host)
		if err == nil {
			host = parsed
		}
	}
	lower := strings.ToLower(host)
	if lower == "0.0.0.0" || lower == "::" || lower == "::1" || lower == "localhost" {
		return ""
	}
	return host
}

func buildInviteAddr(listenAddr, advertiseHost string) string {
	host, port, err := net.SplitHostPort(listenAddr)
	if err != nil {
		return listenAddr
	}
	if forced := sanitizeHostLiteral(advertiseHost); forced != "" {
		return net.JoinHostPort(forced, port)
	}
	chosen := chooseInviteHost(host)
	return net.JoinHostPort(chosen, port)
}

func sanitizeHostLiteral(host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")
	if host == "" {
		return ""
	}
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	}
	return strings.TrimSpace(host)
}

func chooseInviteHost(listenHost string) string {
	listenHost = sanitizeHostLiteral(listenHost)
	if listenHost != "" && !isUnspecifiedHost(listenHost) {
		return listenHost
	}
	if detected := detectAdvertiseHost(); detected != "" {
		return detected
	}
	switch strings.ToLower(listenHost) {
	case "::":
		return "::1"
	default:
		return "127.0.0.1"
	}
}

func isUnspecifiedHost(host string) bool {
	host = sanitizeHostLiteral(host)
	if host == "" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsUnspecified()
}

func detectAdvertiseHost() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	var ipv6Candidate string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip := interfaceAddrIP(addr)
			if ip == nil || !ip.IsGlobalUnicast() || ip.IsLoopback() || ip.IsUnspecified() {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				return ip4.String()
			}
			if !ip.IsLinkLocalUnicast() && ipv6Candidate == "" {
				ipv6Candidate = ip.String()
			}
		}
	}
	return ipv6Candidate
}

func interfaceAddrIP(addr net.Addr) net.IP {
	switch v := addr.(type) {
	case *net.IPNet:
		return v.IP
	case *net.IPAddr:
		return v.IP
	default:
		host, _, err := net.SplitHostPort(addr.String())
		if err == nil {
			return net.ParseIP(host)
		}
		return net.ParseIP(addr.String())
	}
}

type FailoverMetadata struct {
	HostSeat       int
	HandoffPort    int
	PeerHosts      map[int]string
	SeatSessionIDs map[int]string
}

func (h *HostSession) FailoverMetadata() FailoverMetadata {
	h.mu.Lock()
	defer h.mu.Unlock()
	meta := FailoverMetadata{
		HostSeat:       h.tableHostSeat,
		HandoffPort:    h.handoffPort,
		PeerHosts:      make(map[int]string, len(h.peerHosts)),
		SeatSessionIDs: make(map[int]string, len(h.seatID)),
	}
	for seat, host := range h.peerHosts {
		meta.PeerHosts[seat] = host
	}
	for seat, sid := range h.seatID {
		meta.SeatSessionIDs[seat] = sid
	}
	return meta
}

func (h *HostSession) seatAddressFromConnLocked(conn net.Conn, joinAdvertise string) string {
	if normalized := normalizeAdvertiseHost(joinAdvertise); normalized != "" {
		return normalized
	}
	return normalizeAdvertiseHost(remoteIP(conn.RemoteAddr().String()))
}

func (h *HostSession) allowAcceptIPLocked(ip string) bool {
	now := time.Now()
	state := h.acceptRate[ip]
	if state.windowStart.IsZero() || now.Sub(state.windowStart) >= hostAcceptWindow {
		state.windowStart = now
		state.count = 0
	}
	state.count++
	h.acceptRate[ip] = state
	return state.count <= hostAcceptMaxPerIPWindow
}

func (h *HostSession) allowSeatMessageLocked(seat int, kind string) bool {
	now := time.Now()
	st := h.seatRate[seat]
	if st.windowStart.IsZero() || now.Sub(st.windowStart) >= seatRateWindow {
		st.windowStart = now
		st.actions = 0
		st.chat = 0
		st.control = 0
	}
	switch kind {
	case "action":
		st.actions++
		h.seatRate[seat] = st
		return st.actions <= seatMaxActionsPerWindow
	case "chat":
		st.chat++
		h.seatRate[seat] = st
		return st.chat <= seatMaxChatPerWindow
	default:
		st.control++
		h.seatRate[seat] = st
		return st.control <= seatMaxControlPerWindow
	}
}

func (h *HostSession) connectedVoterCountLocked() int {
	count := 1 // seat 0 (processo host) sempre ativo
	for range h.clients {
		count++
	}
	return count
}

func (h *HostSession) transferTableHostLocked(newSeat int, reason string) {
	if newSeat < 0 || newSeat >= h.numPlayers {
		return
	}
	if h.slots[newSeat] == "" {
		return
	}
	old := h.tableHostSeat
	h.tableHostSeat = newSeat
	h.hostVotes = map[int]int{}
	text := fmt.Sprintf("Host da mesa mudou: slot %d (%s).", newSeat+1, h.slots[newSeat])
	if reason != "" {
		text = text + " " + reason
	}
	h.broadcastLocked(Message{Type: "system", Text: text})
	h.sendEventLocked(text)
	if old != newSeat {
		logNetf("table host changed %d -> %d", old, newSeat)
	}
}

func (h *HostSession) castHostVoteLocked(voterSeat, candidateSeat int) error {
	if candidateSeat < 0 || candidateSeat >= h.numPlayers {
		return errors.New("candidato de host inválido")
	}
	if h.slots[candidateSeat] == "" {
		return errors.New("slot alvo vazio")
	}
	if voterSeat != 0 {
		if _, ok := h.clients[voterSeat]; !ok {
			return errors.New("votante desconectado")
		}
	}
	h.hostVotes[voterSeat] = candidateSeat

	votes := 0
	required := h.connectedVoterCountLocked()/2 + 1
	for seat, cand := range h.hostVotes {
		if cand != candidateSeat {
			continue
		}
		if seat == 0 {
			votes++
			continue
		}
		if _, ok := h.clients[seat]; ok {
			votes++
		}
	}
	if votes >= required {
		h.transferTableHostLocked(candidateSeat, "Transferência por votação.")
	}
	return nil
}

func (h *HostSession) maybeWarnTLSExpiryLocked(now time.Time) {
	if h.tlsExpiryWarned || h.tlsNotAfter.IsZero() {
		return
	}
	if now.Add(h.cfg.TLSExpiryWarnBefore).Before(h.tlsNotAfter) {
		return
	}
	remaining := h.tlsNotAfter.Sub(now).Round(time.Minute)
	if remaining < 0 {
		remaining = 0
	}
	text := fmt.Sprintf("Aviso: certificado TLS da sessão expira em %s. Gere novo convite se necessário.", remaining)
	h.tlsExpiryWarned = true
	h.sendEventLocked(text)
	h.broadcastLocked(Message{Type: "system", Text: text})
}

func (h *HostSession) createReplacementInviteLocked(requesterSeat, targetSeat int) (string, error) {
	if !h.started {
		return "", errors.New("convites de reposição disponíveis apenas durante a partida")
	}
	if requesterSeat != h.tableHostSeat {
		return "", errors.New("apenas o host atual da mesa pode gerar convites de reposição")
	}
	if targetSeat <= 0 || targetSeat >= h.numPlayers {
		return "", errors.New("slot de reposição inválido")
	}
	if h.slots[targetSeat] == "" {
		return "", errors.New("slot não ocupado")
	}
	if _, connected := h.clients[targetSeat]; connected {
		return "", errors.New("slot já está conectado")
	}
	replaceToken, err := randomToken()
	if err != nil {
		return "", err
	}
	h.replaceInvites[replaceToken] = targetSeat
	inv := h.inviteBase
	inv.ReplaceToken = replaceToken
	return EncodeInviteKey(inv)
}

func (h *HostSession) sendEventLocked(text string) {
	if h.closed {
		return
	}
	select {
	case h.events <- text:
	default:
	}
}

func (h *HostSession) dropClientLocked(seat int, reason string) {
	c, ok := h.clients[seat]
	if !ok {
		return
	}
	name := h.slots[seat]
	delete(h.clients, seat)
	delete(h.lastPong, seat)
	delete(h.seatRate, seat)
	if !h.started {
		h.slots[seat] = ""
		delete(h.seatID, seat)
	}
	closeConnWithLog(c, "drop client")
	if reason != "" {
		if h.started {
			h.sendEventLocked(fmt.Sprintf("%s desconectou (%s). Aguardando reconexão.", name, reason))
		} else {
			h.sendEventLocked(fmt.Sprintf("%s desconectou (%s).", name, reason))
		}
	}
	if h.tableHostSeat == seat {
		// Transferência automática quando o host da mesa perde conexão.
		if _, ok := h.clients[1]; ok {
			h.transferTableHostLocked(1, "Host anterior desconectou.")
			return
		}
		for candidate := 1; candidate < h.numPlayers; candidate++ {
			if _, ok := h.clients[candidate]; ok {
				h.transferTableHostLocked(candidate, "Host anterior desconectou.")
				return
			}
		}
		h.transferTableHostLocked(0, "Host anterior desconectou.")
	}
}

func (h *HostSession) reconnectSlotLocked(sessionID string) int {
	if strings.TrimSpace(sessionID) == "" {
		return -1
	}
	for i := 1; i < len(h.slots); i++ {
		if h.seatID[i] == sessionID {
			return i
		}
	}
	return -1
}

func (h *HostSession) writeToSeatLocked(seat int, msg Message) bool {
	c, ok := h.clients[seat]
	if !ok {
		return false
	}
	if err := writeMessage(c, msg); err != nil {
		h.dropClientLocked(seat, "erro de escrita")
		return false
	}
	return true
}

func (h *HostSession) broadcastLocked(msg Message) {
	changed := false
	seats := make([]int, 0, len(h.clients))
	for seat := range h.clients {
		seats = append(seats, seat)
	}
	for _, seat := range seats {
		if !h.writeToSeatLocked(seat, msg) {
			changed = true
		}
	}
	if changed && !h.started && msg.Type != "lobby_update" {
		h.broadcastLobbyLocked()
	}
}

func (h *HostSession) broadcastLobbyLocked() {
	h.broadcastLocked(Message{Type: "lobby_update", Slots: append([]string{}, h.slots...), NumPlayers: h.numPlayers})
}

func (h *HostSession) acceptLoop() {
	for {
		conn, err := h.ln.Accept()
		if err != nil {
			if h.ctx.Err() == nil {
				logNetf("accept error: %v", err)
			}
			return
		}
		ip := remoteIP(conn.RemoteAddr().String())
		h.mu.Lock()
		allowed := h.allowAcceptIPLocked(ip)
		h.mu.Unlock()
		if !allowed {
			_ = writeMessage(conn, Message{Type: "error", Error: "muitas conexões, tente novamente"})
			closeConnWithLog(conn, "accept rate limited")
			continue
		}
		go h.handleConn(conn)
	}
}

func (h *HostSession) heartbeatLoop() {
	tk := time.NewTicker(h.cfg.HeartbeatInterval)
	defer tk.Stop()
	for {
		select {
		case <-h.ctx.Done():
			return
		case <-tk.C:
		}
		h.mu.Lock()
		if h.closed {
			h.mu.Unlock()
			return
		}
		now := time.Now()
		h.maybeWarnTLSExpiryLocked(now)
		changed := false
		seats := make([]int, 0, len(h.clients))
		for seat := range h.clients {
			seats = append(seats, seat)
		}
		for _, seat := range seats {
			if _, ok := h.clients[seat]; !ok {
				continue
			}
			last := h.lastPong[seat]
			if !last.IsZero() && now.Sub(last) > h.cfg.HeartbeatTimeout {
				h.dropClientLocked(seat, "heartbeat timeout")
				changed = true
				continue
			}
			if !h.writeToSeatLocked(seat, Message{Type: "ping", HeartbeatUnix: now.Unix()}) {
				changed = true
			}
		}
		if changed && !h.started {
			h.broadcastLobbyLocked()
		}
		h.mu.Unlock()
	}
}

func (h *HostSession) handleConn(conn net.Conn) {
	reader := newConnReader(conn)
	joinMsg, err := readMessage(conn, reader)
	if err != nil || joinMsg.Type != "join" {
		_ = writeMessage(conn, Message{Type: "error", Error: "mensagem inicial inválida"})
		closeConnWithLog(conn, "invalid join")
		return
	}

	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		closeConnWithLog(conn, "host closed before join")
		return
	}
	if joinMsg.ProtocolVersion != protocolVersion {
		h.mu.Unlock()
		_ = writeMessage(conn, Message{Type: "error", Error: fmt.Sprintf("versão de protocolo incompatível (cliente=%d, esperado=%d)", joinMsg.ProtocolVersion, protocolVersion)})
		closeConnWithLog(conn, "protocol version mismatch")
		return
	}
	if joinMsg.Token != h.token {
		h.mu.Unlock()
		_ = writeMessage(conn, Message{Type: "error", Error: "token inválido"})
		closeConnWithLog(conn, "invalid token")
		return
	}
	name, err := normalizeName(joinMsg.Name)
	if err != nil {
		h.mu.Unlock()
		_ = writeMessage(conn, Message{Type: "error", Error: "nome inválido"})
		closeConnWithLog(conn, "invalid name")
		return
	}
	joinMsg.Name = name
	joinMsg.DesiredRole = normalizeDesiredRole(joinMsg.DesiredRole)

	slot := -1
	reconnect := false
	replacement := false
	replacementToken := ""
	previousSlotName := ""
	previousSessionID := ""
	hadPreviousSessionID := false
	var sessionID string
	if h.started {
		slot = h.reconnectSlotLocked(joinMsg.SessionID)
		if slot != -1 {
			if h.slots[slot] != joinMsg.Name {
				h.mu.Unlock()
				_ = writeMessage(conn, Message{Type: "error", Error: "identidade de sessão inválida"})
				closeConnWithLog(conn, "session/name mismatch")
				return
			}
			reconnect = true
		} else if joinMsg.ReplaceToken != "" {
			target, ok := h.replaceInvites[joinMsg.ReplaceToken]
			if !ok {
				h.mu.Unlock()
				_ = writeMessage(conn, Message{Type: "error", Error: "convite de reposição inválido"})
				closeConnWithLog(conn, "invalid replacement invite")
				return
			}
			if _, connected := h.clients[target]; connected {
				h.mu.Unlock()
				_ = writeMessage(conn, Message{Type: "error", Error: "slot já reconectado"})
				closeConnWithLog(conn, "replacement seat already connected")
				return
			}
			slot = target
			replacementToken = joinMsg.ReplaceToken
			previousSlotName = h.slots[slot]
			previousSessionID, hadPreviousSessionID = h.seatID[slot]
			h.slots[slot] = joinMsg.Name
			sessionID, err = randomToken()
			if err != nil {
				h.mu.Unlock()
				_ = writeMessage(conn, Message{Type: "error", Error: "falha ao gerar sessão"})
				closeConnWithLog(conn, "replacement session id generation")
				return
			}
			h.seatID[slot] = sessionID
			replacement = true
		} else {
			h.mu.Unlock()
			_ = writeMessage(conn, Message{Type: "error", Error: "partida em andamento"})
			closeConnWithLog(conn, "join while started")
			return
		}
	} else {
		slot = h.chooseSlotLocked(joinMsg.DesiredRole)
		if slot == -1 {
			h.mu.Unlock()
			_ = writeMessage(conn, Message{Type: "error", Error: "lobby cheio"})
			closeConnWithLog(conn, "lobby full")
			return
		}
		h.slots[slot] = joinMsg.Name
		sessionID, err = randomToken()
		if err != nil {
			h.mu.Unlock()
			_ = writeMessage(conn, Message{Type: "error", Error: "falha ao gerar sessão"})
			closeConnWithLog(conn, "session id generation")
			return
		}
		h.seatID[slot] = sessionID
	}
	if old, ok := h.clients[slot]; ok && old != nil {
		closeConnWithLog(old, "replace seat connection")
	}
	h.clients[slot] = conn
	h.peerHosts[slot] = h.seatAddressFromConnLocked(conn, joinMsg.AdvertiseHost)
	h.lastPong[slot] = time.Now()
	cachedState, hasCachedState := h.lastState[slot]
	if err := writeMessage(conn, Message{Type: "join_ok", ProtocolVersion: protocolVersion, Assigned: slot, NumPlayers: h.numPlayers, Slots: append([]string{}, h.slots...), SessionID: h.seatID[slot]}); err != nil {
		if replacement {
			h.slots[slot] = previousSlotName
			if hadPreviousSessionID {
				h.seatID[slot] = previousSessionID
			} else {
				delete(h.seatID, slot)
			}
			if replacementToken != "" {
				h.replaceInvites[replacementToken] = slot
			}
		}
		h.dropClientLocked(slot, "falha no handshake")
		h.mu.Unlock()
		closeConnWithLog(conn, "join ack failure")
		return
	}
	if replacementToken != "" {
		delete(h.replaceInvites, replacementToken)
	}
	if reconnect && hasCachedState {
		snap := cloneSnapshot(cachedState)
		if err := writeMessage(conn, Message{Type: "game_state", State: &snap}); err != nil {
			h.dropClientLocked(slot, "falha ao sincronizar estado")
			h.mu.Unlock()
			closeConnWithLog(conn, "reconnect state sync failure")
			return
		}
	}
	if !h.started {
		h.broadcastLobbyLocked()
		h.sendEventLocked(fmt.Sprintf("%s entrou no slot %d.", joinMsg.Name, slot+1))
	} else if replacement {
		h.sendEventLocked(fmt.Sprintf("%s entrou no slot %d como substituto (CPU provisório removido).", joinMsg.Name, slot+1))
		h.broadcastLocked(Message{Type: "system", Text: fmt.Sprintf("%s assumiu o slot %d.", joinMsg.Name, slot+1)})
	} else {
		h.sendEventLocked(fmt.Sprintf("%s reconectou no slot %d.", joinMsg.Name, slot+1))
	}
	h.mu.Unlock()

	for {
		select {
		case <-h.ctx.Done():
			return
		default:
		}
		msg, err := readMessage(conn, reader)
		if err != nil {
			break
		}
		switch msg.Type {
		case "chat":
			h.mu.Lock()
			allowed := h.allowSeatMessageLocked(slot, "chat")
			h.mu.Unlock()
			if !allowed {
				h.SendSystemToSeat(slot, "Rate limit de chat excedido.")
				continue
			}
			text, err := normalizeChatText(msg.Text)
			if err != nil {
				h.SendSystemToSeat(slot, "Mensagem inválida.")
				continue
			}
			h.mu.Lock()
			from := h.slots[slot]
			h.broadcastLocked(Message{Type: "chat", Name: from, Text: text})
			h.sendEventLocked(fmt.Sprintf("[chat] %s: %s", from, text))
			h.mu.Unlock()
		case "game_action":
			h.mu.Lock()
			allowed := h.allowSeatMessageLocked(slot, "action")
			h.mu.Unlock()
			if !allowed {
				h.SendSystemToSeat(slot, "Rate limit de ações excedido.")
				continue
			}
			if err := validateGameAction(msg.Action, msg.CardIndex); err != nil {
				h.SendSystemToSeat(slot, "Ação inválida.")
				continue
			}
			h.mu.Lock()
			started := h.started && !h.closed
			h.mu.Unlock()
			if !started {
				continue
			}
			action := ClientAction{Seat: slot, Action: msg.Action, CardIndex: msg.CardIndex}
			select {
			case h.actions <- action:
			case <-time.After(2 * time.Second):
				h.SendSystemToSeat(slot, "Host congestionado: ação não processada a tempo.")
				h.mu.Lock()
				h.sendEventLocked(fmt.Sprintf("[warn] ação descartada por timeout do slot %d", slot+1))
				h.mu.Unlock()
			}
		case "host_vote":
			h.mu.Lock()
			allowed := h.allowSeatMessageLocked(slot, "control")
			if !allowed {
				h.mu.Unlock()
				h.SendSystemToSeat(slot, "Rate limit de controle excedido.")
				continue
			}
			err := h.castHostVoteLocked(slot, msg.HostCandidate)
			h.mu.Unlock()
			if err != nil {
				h.SendSystemToSeat(slot, "Voto inválido: "+err.Error())
			}
		case "invite_request":
			h.mu.Lock()
			allowed := h.allowSeatMessageLocked(slot, "control")
			if !allowed {
				h.mu.Unlock()
				h.SendSystemToSeat(slot, "Rate limit de controle excedido.")
				continue
			}
			key, err := h.createReplacementInviteLocked(slot, msg.TargetSeat)
			if err != nil {
				h.writeToSeatLocked(slot, Message{Type: "system", Text: "Falha ao gerar convite: " + err.Error()})
				h.mu.Unlock()
				continue
			}
			h.writeToSeatLocked(slot, Message{Type: "system", Text: fmt.Sprintf("Convite de reposição (slot %d): %s", msg.TargetSeat+1, key)})
			h.mu.Unlock()
		case "ping":
			h.mu.Lock()
			h.lastPong[slot] = time.Now()
			h.writeToSeatLocked(slot, Message{Type: "pong", HeartbeatUnix: time.Now().Unix()})
			h.mu.Unlock()
		case "pong":
			h.mu.Lock()
			h.lastPong[slot] = time.Now()
			h.mu.Unlock()
		}
	}

	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		closeConnWithLog(conn, "disconnect after host close")
		return
	}
	current, ok := h.clients[slot]
	if !ok || current != conn {
		h.mu.Unlock()
		closeConnWithLog(conn, "stale connection")
		return
	}
	name = h.slots[slot]
	h.dropClientLocked(slot, "conexão perdida")
	if !h.started {
		h.broadcastLobbyLocked()
		h.sendEventLocked(fmt.Sprintf("%s saiu da sessão.", name))
	}
	h.mu.Unlock()
}

func (h *HostSession) SendHostChat(text string) {
	text, err := normalizeChatText(text)
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.broadcastLocked(Message{Type: "chat", Name: h.hostName, Text: text})
	h.sendEventLocked(fmt.Sprintf("[chat] %s: %s", h.hostName, text))
}

func (h *HostSession) SendSystem(text string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.broadcastLocked(Message{Type: "system", Text: text})
	h.sendEventLocked("[system] " + text)
}

func (h *HostSession) SendSystemToSeat(seat int, text string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	ok := h.writeToSeatLocked(seat, Message{Type: "system", Text: text})
	if !ok && !h.started {
		h.broadcastLobbyLocked()
	}
}

func (h *HostSession) SendGameStateToSeat(seat int, state Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	if state.Type == "game_state" {
		state.HostSeat = h.tableHostSeat
		state.HandoffPort = h.handoffPort
		state.PeerHosts = make(map[int]string, len(h.peerHosts))
		for k, v := range h.peerHosts {
			state.PeerHosts[k] = v
		}
		state.SeatSessionIDs = make(map[int]string, len(h.seatID))
		for k, v := range h.seatID {
			state.SeatSessionIDs[k] = v
		}
		if state.State != nil {
			snap := trimSnapshotForWire(*state.State, scannerMaxBuffer)
			h.lastState[seat] = snap
			state.State = &snap
		}
		if state.FullState != nil {
			full := trimSnapshotForFailover(*state.FullState)
			state.FullState = &full
		}
	}
	ok := h.writeToSeatLocked(seat, state)
	if !ok && !h.started {
		h.broadcastLobbyLocked()
	}
}
