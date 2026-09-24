package netp2p

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"tailscale.com/tsnet"
)

const (
	defaultTailnetRequestTimeout = 8 * time.Second
	envCoordinatorURL            = "TRUCO_COORDINATOR_URL"
	envTailnetControlURL         = "TRUCO_TAILNET_CONTROL_URL"
	envTailnetAuthKey            = "TRUCO_TAILNET_AUTH_KEY"
	envTailnetStateDir           = "TRUCO_TAILNET_STATE_DIR"
)

type TailnetCoordinator interface {
	CreateTailnetSession(context.Context, TailnetCreateSessionRequest) (TailnetCreateSessionResponse, error)
	JoinTailnetSession(context.Context, TailnetJoinSessionRequest) (TailnetJoinSessionResponse, error)
	MintTailnetJoinTicket(context.Context, TailnetMintJoinTicketRequest) (TailnetMintJoinTicketResponse, error)
	PromoteTailnetAuthority(context.Context, TailnetPromoteAuthorityRequest) (TailnetPromoteAuthorityResponse, error)
}

type TailnetCreateSessionRequest struct {
	HostIdentity string `json:"host_identity"`
	NumPlayers   int    `json:"num_players"`
	ServicePort  int    `json:"service_port"`
}

type TailnetCreateSessionResponse struct {
	SessionID      string    `json:"session_id"`
	HostAdminToken string    `json:"host_admin_token"`
	HostAuthKey    string    `json:"host_auth_key"`
	HostNodeName   string    `json:"host_node_name"`
	ControlURL     string    `json:"control_url,omitempty"`
	JoinTicket     string    `json:"join_ticket"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type TailnetJoinSessionRequest struct {
	SessionID     string `json:"session_id"`
	JoinTicket    string `json:"join_ticket"`
	PlayerName    string `json:"player_name"`
	DesiredRole   string `json:"desired_role,omitempty"`
	PlayerSession string `json:"player_session,omitempty"`
}

type TailnetJoinSessionResponse struct {
	AuthKey       string    `json:"auth_key"`
	NodeName      string    `json:"node_name"`
	AuthorityNode string    `json:"authority_node"`
	ControlURL    string    `json:"control_url,omitempty"`
	ServicePort   int       `json:"service_port"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type TailnetMintJoinTicketRequest struct {
	SessionID      string `json:"session_id"`
	HostAdminToken string `json:"host_admin_token"`
	PlayerName     string `json:"player_name,omitempty"`
	DesiredRole    string `json:"desired_role,omitempty"`
	TargetSeat     int    `json:"target_seat,omitempty"`
}

type TailnetMintJoinTicketResponse struct {
	JoinTicket string    `json:"join_ticket"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type TailnetPromoteAuthorityRequest struct {
	SessionID      string `json:"session_id"`
	HostAdminToken string `json:"host_admin_token"`
	HostIdentity   string `json:"host_identity"`
	AuthorityNode  string `json:"authority_node,omitempty"`
	ServicePort    int    `json:"service_port"`
	Epoch          int    `json:"epoch"`
}

type TailnetPromoteAuthorityResponse struct {
	HostAuthKey   string    `json:"host_auth_key"`
	AuthorityNode string    `json:"authority_node"`
	ControlURL    string    `json:"control_url,omitempty"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type TransportDiagnostics struct {
	RequestedTransport string
	SelectedTransport  string
	DirectPathKnown    bool
	DirectPath         bool
	RelayFallback      bool
	CoordinatorStatus  string
	CoordinatorURL     string
	TailnetNode        string
	TailnetAuthority   string
	TailnetServicePort int
	FallbackReason     string
}

type tailnetEndpoint struct {
	server            *tsnet.Server
	listener          net.Listener
	nodeName          string
	coordinatorURL    string
	controlURL        string
	sessionID         string
	hostAdminToken    string
	joinTicket        string
	expiresAt         time.Time
	authorityNode     string
	servicePort       int
	coordinatorStatus string
}

func (e *tailnetEndpoint) Close() error {
	var err error
	if e.listener != nil {
		err = e.listener.Close()
	}
	if e.server != nil {
		if closeErr := e.server.Close(); err == nil {
			err = closeErr
		}
	}
	return err
}

type tailnetConn struct {
	net.Conn
	endpoint *tailnetEndpoint
}

func (c *tailnetConn) Close() error {
	err := c.Conn.Close()
	if c.endpoint != nil {
		if closeErr := c.endpoint.Close(); err == nil {
			err = closeErr
		}
	}
	return err
}

func hasTailnetConfig(c HostConfig) bool {
	return c.TailnetCoordinator != nil ||
		strings.TrimSpace(c.TailnetCoordinatorURL) != "" ||
		strings.TrimSpace(os.Getenv(envCoordinatorURL)) != ""
}

func startTailnetHost(ctx context.Context, cfg HostConfig, hostName string, numPlayers int) (*tailnetEndpoint, error) {
	coordinatorURL := firstNonEmpty(cfg.TailnetCoordinatorURL, os.Getenv(envCoordinatorURL))
	controlURL := firstNonEmpty(cfg.TailnetControlURL, os.Getenv(envTailnetControlURL))
	servicePort := cfg.TailnetServicePort
	if servicePort <= 0 {
		servicePort = cfg.HandoffPort
	}
	nodeName := firstNonEmpty(cfg.TailnetNodeName, "truco-host-"+shortRandomLabel())
	authKey := firstNonEmpty(cfg.TailnetAuthKey, os.Getenv(envTailnetAuthKey))
	sessionID := strings.TrimSpace(cfg.TailnetSessionID)
	hostAdminToken := strings.TrimSpace(cfg.TailnetHostAdminToken)
	expiresAt := time.Now().UTC().Add(5 * time.Minute)
	joinTicket := ""

	coordinator := cfg.TailnetCoordinator
	if coordinator == nil && coordinatorURL != "" {
		coordinator = newHTTPTailnetCoordinator(coordinatorURL)
	}
	if coordinator != nil && authKey == "" {
		var err error
		if sessionID != "" && hostAdminToken != "" {
			promoted, promoteErr := coordinator.PromoteTailnetAuthority(ctx, TailnetPromoteAuthorityRequest{
				SessionID:      sessionID,
				HostAdminToken: hostAdminToken,
				HostIdentity:   hostName,
				AuthorityNode:  nodeName,
				ServicePort:    servicePort,
				Epoch:          maxInt(1, cfg.RelayEpoch),
			})
			if promoteErr != nil {
				return nil, promoteErr
			}
			authKey = strings.TrimSpace(promoted.HostAuthKey)
			nodeName = firstNonEmpty(promoted.AuthorityNode, nodeName)
			controlURL = firstNonEmpty(promoted.ControlURL, controlURL)
			expiresAt = nonZeroTime(promoted.ExpiresAt, expiresAt)
		} else {
			created, createErr := coordinator.CreateTailnetSession(ctx, TailnetCreateSessionRequest{
				HostIdentity: hostName,
				NumPlayers:   numPlayers,
				ServicePort:  servicePort,
			})
			if createErr != nil {
				return nil, createErr
			}
			sessionID = strings.TrimSpace(created.SessionID)
			hostAdminToken = strings.TrimSpace(created.HostAdminToken)
			authKey = strings.TrimSpace(created.HostAuthKey)
			nodeName = firstNonEmpty(created.HostNodeName, nodeName)
			controlURL = firstNonEmpty(created.ControlURL, controlURL)
			joinTicket = strings.TrimSpace(created.JoinTicket)
			expiresAt = nonZeroTime(created.ExpiresAt, expiresAt)
		}
		if authKey == "" {
			err = errors.New("coordenador tailnet não retornou auth key")
		}
		if err != nil {
			return nil, err
		}
	}
	if authKey == "" {
		return nil, errors.New("tailnet requer coordenador ou auth key")
	}

	stateDir, err := tailnetStateDir(cfg.TailnetStateDir, nodeName)
	if err != nil {
		return nil, err
	}
	server := &tsnet.Server{
		Hostname:   nodeName,
		Dir:        stateDir,
		AuthKey:    authKey,
		ControlURL: controlURL,
		Ephemeral:  true,
		Logf:       logNetf,
	}
	listener, err := server.Listen("tcp", net.JoinHostPort("", strconv.Itoa(servicePort)))
	if err != nil {
		_ = server.Close()
		return nil, err
	}
	actualPort := listenerPort(listener, servicePort)
	ep := &tailnetEndpoint{
		server:            server,
		listener:          listener,
		nodeName:          nodeName,
		coordinatorURL:    coordinatorURL,
		controlURL:        controlURL,
		sessionID:         sessionID,
		hostAdminToken:    hostAdminToken,
		joinTicket:        joinTicket,
		expiresAt:         expiresAt,
		authorityNode:     nodeName,
		servicePort:       actualPort,
		coordinatorStatus: "connected",
	}
	return ep, nil
}

func tailnetCoordinatorFromConfig(cfg HostConfig) TailnetCoordinator {
	if cfg.TailnetCoordinator != nil {
		return cfg.TailnetCoordinator
	}
	if coordinatorURL := firstNonEmpty(cfg.TailnetCoordinatorURL, os.Getenv(envCoordinatorURL)); coordinatorURL != "" {
		return newHTTPTailnetCoordinator(coordinatorURL)
	}
	return nil
}

func mintTailnetJoinTicket(ctx context.Context, cfg HostConfig, playerName, desiredRole string, targetSeat int) (TailnetMintJoinTicketResponse, error) {
	coordinator := tailnetCoordinatorFromConfig(cfg)
	if coordinator == nil {
		return TailnetMintJoinTicketResponse{}, errors.New("coordenador tailnet ausente")
	}
	sessionID := strings.TrimSpace(cfg.TailnetSessionID)
	hostAdminToken := strings.TrimSpace(cfg.TailnetHostAdminToken)
	if sessionID == "" || hostAdminToken == "" {
		return TailnetMintJoinTicketResponse{}, errors.New("sessão tailnet sem credencial administrativa")
	}
	return coordinator.MintTailnetJoinTicket(ctx, TailnetMintJoinTicketRequest{
		SessionID:      sessionID,
		HostAdminToken: hostAdminToken,
		PlayerName:     playerName,
		DesiredRole:    desiredRole,
		TargetSeat:     targetSeat,
	})
}

func dialTailnetSessionConn(inv InviteKey, timeout time.Duration, playerName, desiredRole, playerSession string, cachedState *RelayReconnectState) (net.Conn, *RelayReconnectState, error) {
	if cachedState != nil && strings.TrimSpace(cachedState.TailnetNodeName) != "" {
		conn, state, err := dialTailnetWithNodeState(inv, timeout, cachedState)
		if err == nil {
			return conn, state, nil
		}
		logNetf("tailnet cached reconnect failed: %v", err)
	}
	coordinatorURL := strings.TrimSpace(inv.TailnetCoordinatorURL)
	if coordinatorURL == "" {
		return nil, nil, errors.New("coordenador tailnet ausente")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	joined, err := newHTTPTailnetCoordinator(coordinatorURL).JoinTailnetSession(ctx, TailnetJoinSessionRequest{
		SessionID:     inv.TailnetSessionID,
		JoinTicket:    inv.TailnetJoinTicket,
		PlayerName:    playerName,
		DesiredRole:   desiredRole,
		PlayerSession: playerSession,
	})
	if err != nil {
		return nil, nil, err
	}
	nodeName := firstNonEmpty(joined.NodeName, "truco-client-"+shortRandomLabel())
	controlURL := firstNonEmpty(joined.ControlURL, inv.TailnetControlURL, os.Getenv(envTailnetControlURL))
	servicePort := joined.ServicePort
	if servicePort <= 0 {
		servicePort = inv.TailnetServicePort
	}
	authorityNode := firstNonEmpty(joined.AuthorityNode, inv.TailnetAuthorityNode)
	if strings.TrimSpace(joined.AuthKey) == "" || authorityNode == "" || servicePort <= 0 {
		return nil, nil, errors.New("coordenador tailnet retornou dados incompletos")
	}
	state := &RelayReconnectState{
		TailnetNodeName:     nodeName,
		TailnetControlURL:   controlURL,
		TailnetCoordinator:  coordinatorURL,
		TailnetAuthority:    authorityNode,
		TailnetServicePort:  servicePort,
		TailnetConnectedP2P: true,
	}
	return dialTailnetWithNodeState(inv, timeout, state, joined.AuthKey)
}

func dialTailnetWithNodeState(inv InviteKey, timeout time.Duration, state *RelayReconnectState, authKey ...string) (net.Conn, *RelayReconnectState, error) {
	if state == nil {
		return nil, nil, errors.New("estado tailnet ausente")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	nodeName := strings.TrimSpace(state.TailnetNodeName)
	controlURL := firstNonEmpty(state.TailnetControlURL, inv.TailnetControlURL, os.Getenv(envTailnetControlURL))
	coordinatorURL := firstNonEmpty(state.TailnetCoordinator, inv.TailnetCoordinatorURL)
	authorityNode := firstNonEmpty(state.TailnetAuthority, inv.TailnetAuthorityNode)
	servicePort := firstNonZero(state.TailnetServicePort, inv.TailnetServicePort)
	if nodeName == "" || authorityNode == "" || servicePort <= 0 {
		return nil, nil, errors.New("estado tailnet incompleto")
	}
	stateDir, err := tailnetStateDir("", nodeName)
	if err != nil {
		return nil, nil, err
	}
	key := ""
	if len(authKey) > 0 {
		key = strings.TrimSpace(authKey[0])
	}
	server := &tsnet.Server{
		Hostname:   nodeName,
		Dir:        stateDir,
		AuthKey:    key,
		ControlURL: controlURL,
		Ephemeral:  true,
		Logf:       logNetf,
	}
	raw, err := server.Dial(ctx, "tcp", net.JoinHostPort(authorityNode, strconv.Itoa(servicePort)))
	if err != nil {
		_ = server.Close()
		return nil, nil, err
	}
	wrapped, err := wrapTLSClient(raw, inv)
	if err != nil {
		_ = server.Close()
		return nil, nil, err
	}
	out := *state
	out.TailnetControlURL = controlURL
	out.TailnetAuthority = authorityNode
	out.TailnetServicePort = servicePort
	out.TailnetConnectedP2P = true
	return &tailnetConn{Conn: wrapped, endpoint: &tailnetEndpoint{server: server, nodeName: nodeName, coordinatorURL: coordinatorURL, controlURL: controlURL, authorityNode: authorityNode, servicePort: servicePort, coordinatorStatus: "connected"}}, &out, nil
}

type httpTailnetCoordinator struct {
	baseURL string
	client  *http.Client
}

func newHTTPTailnetCoordinator(baseURL string) *httpTailnetCoordinator {
	return &httpTailnetCoordinator{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		client:  &http.Client{Timeout: defaultTailnetRequestTimeout},
	}
}

func (c *httpTailnetCoordinator) CreateTailnetSession(ctx context.Context, req TailnetCreateSessionRequest) (TailnetCreateSessionResponse, error) {
	var out TailnetCreateSessionResponse
	err := c.post(ctx, "/v1/tailnet/sessions", req, &out)
	return out, err
}

func (c *httpTailnetCoordinator) JoinTailnetSession(ctx context.Context, req TailnetJoinSessionRequest) (TailnetJoinSessionResponse, error) {
	var out TailnetJoinSessionResponse
	err := c.post(ctx, "/v1/tailnet/sessions/"+url.PathEscape(req.SessionID)+"/join", req, &out)
	return out, err
}

func (c *httpTailnetCoordinator) MintTailnetJoinTicket(ctx context.Context, req TailnetMintJoinTicketRequest) (TailnetMintJoinTicketResponse, error) {
	var out TailnetMintJoinTicketResponse
	err := c.post(ctx, "/v1/tailnet/sessions/"+url.PathEscape(req.SessionID)+"/join-tickets", req, &out)
	return out, err
}

func (c *httpTailnetCoordinator) PromoteTailnetAuthority(ctx context.Context, req TailnetPromoteAuthorityRequest) (TailnetPromoteAuthorityResponse, error) {
	var out TailnetPromoteAuthorityResponse
	err := c.post(ctx, "/v1/tailnet/sessions/"+url.PathEscape(req.SessionID)+"/authority", req, &out)
	return out, err
}

func (c *httpTailnetCoordinator) post(ctx context.Context, route string, in any, out any) error {
	if c.baseURL == "" {
		return errors.New("coordenador tailnet ausente")
	}
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+route, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var payload map[string]any
		_ = json.NewDecoder(res.Body).Decode(&payload)
		if msg, ok := payload["error"].(string); ok && msg != "" {
			return fmt.Errorf("coordenador tailnet: %s", msg)
		}
		return fmt.Errorf("coordenador tailnet status %d", res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(out)
}

func tailnetStateDir(configured, nodeName string) (string, error) {
	if configured == "" {
		configured = os.Getenv(envTailnetStateDir)
	}
	if configured == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		configured = filepath.Join(base, "truco-tui", "tailnet")
	}
	dir := filepath.Join(configured, safeTailnetNodeName(nodeName))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

func listenerPort(listener net.Listener, fallback int) int {
	if listener == nil {
		return fallback
	}
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		return fallback
	}
	parsed, err := strconv.Atoi(port)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func safeTailnetNodeName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "truco-node"
	}
	return out
}

func shortRandomLabel() string {
	token, err := randomToken()
	if err != nil || len(token) < 8 {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return token[:8]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func nonZeroTime(value time.Time, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback
	}
	return value
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
