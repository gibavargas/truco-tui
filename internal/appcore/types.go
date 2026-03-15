package appcore

import (
	"encoding/json"

	"truco-tui/internal/truco"
)

const (
	CoreAPIVersion      = 1
	SnapshotSchemaMajor = 1
)

type AppIntent struct {
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type AppEvent struct {
	Kind      string      `json:"kind"`
	Sequence  int64       `json:"sequence"`
	Timestamp string      `json:"timestamp"`
	Payload   interface{} `json:"payload,omitempty"`
}

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CoreVersions struct {
	CoreAPIVersion  int `json:"core_api_version"`
	ProtocolVersion int `json:"protocol_version"`
	SnapshotSchema  int `json:"snapshot_schema_version"`
}

type LobbySnapshot struct {
	InviteKey      string         `json:"invite_key,omitempty"`
	Slots          []string       `json:"slots,omitempty"`
	AssignedSeat   int            `json:"assigned_seat"`
	NumPlayers     int            `json:"num_players"`
	Started        bool           `json:"started"`
	HostSeat       int            `json:"host_seat"`
	ConnectedSeats map[int]bool   `json:"connected_seats,omitempty"`
	Role           string         `json:"role,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

type LobbySlotState struct {
	Seat                  int    `json:"seat"`
	Name                  string `json:"name,omitempty"`
	Status                string `json:"status"`
	IsEmpty               bool   `json:"is_empty"`
	IsLocal               bool   `json:"is_local"`
	IsHost                bool   `json:"is_host"`
	IsConnected           bool   `json:"is_connected"`
	IsOccupied            bool   `json:"is_occupied"`
	IsProvisionalCPU      bool   `json:"is_provisional_cpu"`
	CanVoteHost           bool   `json:"can_vote_host"`
	CanRequestReplacement bool   `json:"can_request_replacement"`
}

type ActionSnapshot struct {
	LocalPlayerID   int  `json:"local_player_id"`
	LocalTeam       int  `json:"local_team"`
	CanPlayCard     bool `json:"can_play_card"`
	CanAskOrRaise   bool `json:"can_ask_or_raise"`
	MustRespond     bool `json:"must_respond"`
	CanAccept       bool `json:"can_accept"`
	CanRefuse       bool `json:"can_refuse"`
	CanCloseSession bool `json:"can_close_session"`
}

type UIStateSnapshot struct {
	LobbySlots []LobbySlotState `json:"lobby_slots,omitempty"`
	Actions    ActionSnapshot   `json:"actions"`
}

type ConnectionSnapshot struct {
	Status       string    `json:"status"`
	IsOnline     bool      `json:"is_online"`
	IsHost       bool      `json:"is_host"`
	LastError    *AppError `json:"last_error,omitempty"`
	LastEventSeq int64     `json:"last_event_sequence"`
}

type DiagnosticsSnapshot struct {
	EventBacklog int      `json:"event_backlog"`
	ReplaySeedLo uint64   `json:"replay_seed_lo,omitempty"`
	ReplaySeedHi uint64   `json:"replay_seed_hi,omitempty"`
	EventLog     []string `json:"event_log,omitempty"`
}

type SnapshotBundle struct {
	Versions    CoreVersions        `json:"versions"`
	Mode        string              `json:"mode"`
	Locale      string              `json:"locale"`
	Match       *truco.Snapshot     `json:"match,omitempty"`
	Lobby       *LobbySnapshot      `json:"lobby,omitempty"`
	UI          UIStateSnapshot     `json:"ui"`
	Connection  ConnectionSnapshot  `json:"connection"`
	Diagnostics DiagnosticsSnapshot `json:"diagnostics"`
}

type SetLocalePayload struct {
	Locale string `json:"locale"`
}

type NewOfflineGamePayload struct {
	PlayerNames []string `json:"player_names"`
	CPUFlags    []bool   `json:"cpu_flags"`
	SeedLo      uint64   `json:"seed_lo,omitempty"`
	SeedHi      uint64   `json:"seed_hi,omitempty"`
}

type CreateHostPayload struct {
	BindAddr      string `json:"bind_addr,omitempty"`
	HostName      string `json:"host_name"`
	NumPlayers    int    `json:"num_players"`
	RelayURL      string `json:"relay_url,omitempty"`
	TransportMode string `json:"transport_mode,omitempty"`
}

type JoinSessionPayload struct {
	Key         string `json:"key"`
	PlayerName  string `json:"player_name"`
	DesiredRole string `json:"desired_role,omitempty"`
}

type GameActionPayload struct {
	Action    string `json:"action"`
	CardIndex int    `json:"card_index,omitempty"`
}

type SendChatPayload struct {
	Text string `json:"text"`
}

type HostVotePayload struct {
	CandidateSeat int `json:"candidate_seat"`
}

type ReplacementInvitePayload struct {
	TargetSeat int `json:"target_seat"`
}
