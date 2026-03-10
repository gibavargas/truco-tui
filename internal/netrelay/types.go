package netrelay

import "time"

const (
	TunnelProto = "truco-relay-quic-v2"
)

type CreateSessionRequest struct {
	HostIdentity string `json:"host_identity"`
	NumPlayers   int    `json:"num_players"`
}

type CreateSessionResponse struct {
	SessionID          string    `json:"session_id"`
	HostAdminToken     string    `json:"host_admin_token"`
	HostPeerID         string    `json:"host_peer_id"`
	HostPeerCredential string    `json:"host_peer_credential"`
	AuthorityPeerID    string    `json:"authority_peer_id"`
	Epoch              int       `json:"epoch"`
	QuicAddr           string    `json:"quic_addr"`
	ExpiresAt          time.Time `json:"expires_at"`
}

type MintJoinTicketRequest struct {
	SessionID      string `json:"session_id"`
	HostAdminToken string `json:"host_admin_token"`
	PlayerName     string `json:"player_name"`
	DesiredRole    string `json:"desired_role,omitempty"`
	TargetSeat     int    `json:"target_seat,omitempty"`
	PlayerSession  string `json:"player_session,omitempty"`
}

type MintJoinTicketResponse struct {
	JoinTicket string    `json:"join_ticket"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type JoinSessionRequest struct {
	SessionID    string `json:"session_id"`
	JoinTicket   string `json:"join_ticket"`
	PlayerName   string `json:"player_name"`
	DesiredRole  string `json:"desired_role,omitempty"`
	TargetSeat   int    `json:"target_seat,omitempty"`
	PlayerSession string `json:"player_session,omitempty"`
}

type JoinSessionResponse struct {
	PeerID             string    `json:"peer_id"`
	PeerCredential     string    `json:"peer_credential"`
	AuthorityPeerID    string    `json:"authority_peer_id"`
	Epoch              int       `json:"epoch"`
	QuicAddr           string    `json:"quic_addr"`
	ExpiresAt          time.Time `json:"expires_at"`
	SessionExpiresAt   time.Time `json:"session_expires_at"`
}

type PublishAuthorityRequest struct {
	SessionID      string `json:"session_id"`
	HostAdminToken string `json:"host_admin_token"`
	HostPeerID     string `json:"host_peer_id"`
	HostIdentity   string `json:"host_identity"`
	Epoch          int    `json:"epoch"`
	RouteHint      string `json:"route_hint,omitempty"`
	ReconnectToken string `json:"reconnect_token,omitempty"`
}

type PublishAuthorityResponse struct {
	AuthorityPeerID    string `json:"authority_peer_id"`
	Epoch              int    `json:"epoch"`
	HostPeerCredential string `json:"host_peer_credential"`
	QuicAddr           string `json:"quic_addr"`
}

type HeartbeatRequest struct {
	SessionID      string `json:"session_id"`
	PeerID         string `json:"peer_id"`
	PeerCredential string `json:"peer_credential"`
	Epoch          int    `json:"epoch"`
}

type tunnelHello struct {
	Type         string `json:"type"`
	SessionID    string `json:"session_id"`
	PeerID       string `json:"peer_id"`
	Credential   string `json:"credential"`
	TargetPeerID string `json:"target_peer_id,omitempty"`
}
