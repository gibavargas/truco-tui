package netp2p

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"truco-tui/internal/truco"
)

var netp2pLogger = log.New(os.Stderr, "[netp2p] ", log.LstdFlags)

const (
	protocolVersion = 2

	writeMessageDeadline = 3 * time.Second
	readMessageDeadline  = 45 * time.Second

	scannerInitialBuffer = 1024
	scannerMaxBuffer     = 16 * 1024

	maxPlayerNameLen = 32
	maxChatTextLen   = 256
)

const ProtocolVersion = protocolVersion

func logNetf(format string, args ...any) {
	if netp2pLogger != nil {
		netp2pLogger.Printf(format, args...)
	}
}

func closeConnWithLog(conn net.Conn, context string) {
	if conn == nil {
		return
	}
	if err := conn.Close(); err != nil {
		logNetf("close conn (%s): %v", context, err)
	}
}

// InviteKey carrega os dados mínimos para conectar a um host.
type InviteKey struct {
	Addr               string `json:"addr,omitempty"`
	Token              string `json:"token"`
	Fingerprint        string `json:"fingerprint,omitempty"`
	ReplaceToken       string `json:"replace_token,omitempty"`
	Transport          string `json:"transport,omitempty"` // tcp_tls|relay_quic_v2
	TransportVersion   int    `json:"transport_version,omitempty"`
	RelayURL           string `json:"relay_url,omitempty"`
	RelaySessionID     string `json:"relay_session_id,omitempty"`
	RelayJoinTicket    string `json:"relay_join_ticket,omitempty"`
	RelayAuthorityPeer string `json:"relay_authority_peer,omitempty"`
	RelaySPKIPin       string `json:"relay_spki_pin,omitempty"`
	ExpiresAt          string `json:"expires_at,omitempty"`
}

func EncodeInviteKey(k InviteKey) (string, error) {
	b, err := json.Marshal(k)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func DecodeInviteKey(s string) (InviteKey, error) {
	var k InviteKey
	b, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil {
		return k, err
	}
	if err := json.Unmarshal(b, &k); err != nil {
		return k, err
	}
	if k.Token == "" {
		return k, errors.New("chave inválida")
	}
	if k.TransportVersion != 2 {
		return k, errors.New("chave de convite v1 não suportada; atualize para v2")
	}
	if strings.TrimSpace(k.Fingerprint) == "" {
		return k, errors.New("chave inválida: fingerprint TLS obrigatório")
	}
	if strings.TrimSpace(k.ExpiresAt) != "" {
		exp, err := time.Parse(time.RFC3339, k.ExpiresAt)
		if err != nil {
			return k, errors.New("chave inválida: expires_at")
		}
		if time.Now().After(exp) {
			return k, errors.New("chave expirada")
		}
	}
	switch strings.TrimSpace(k.Transport) {
	case "tcp_tls":
		if k.Addr == "" {
			return k, errors.New("chave inválida")
		}
	case "relay_quic_v2":
		if strings.TrimSpace(k.RelayURL) == "" || strings.TrimSpace(k.RelaySessionID) == "" || strings.TrimSpace(k.RelayJoinTicket) == "" {
			return k, errors.New("chave de relay inválida")
		}
	default:
		return k, errors.New("transporte de convite inválido")
	}
	return k, nil
}

// Message é o protocolo JSON linha-a-linha da sessão P2P.
//
// Tipos principais:
// - lobby: join, join_ok, lobby_update, chat
// - partida: game_start, game_action, game_state, system
// - erro: error
type Message struct {
	Type            string   `json:"type"`
	ProtocolVersion int      `json:"protocol_version,omitempty"`
	Token           string   `json:"token,omitempty"`
	Name            string   `json:"name,omitempty"`
	SessionID       string   `json:"session_id,omitempty"`
	ReplaceToken    string   `json:"replace_token,omitempty"`
	DesiredRole     string   `json:"desired_role,omitempty"` // partner|opponent|auto
	AdvertiseHost   string   `json:"advertise_host,omitempty"`
	Text            string   `json:"text,omitempty"`
	Slots           []string `json:"slots,omitempty"`
	Assigned        int      `json:"assigned"`
	NumPlayers      int      `json:"num_players,omitempty"`
	Error           string   `json:"error,omitempty"`

	// Campos de ação de partida (cliente -> host).
	Action        string `json:"action,omitempty"` // play|truco|accept|refuse
	CardIndex     int    `json:"card_index"`
	HostCandidate int    `json:"host_candidate,omitempty"`
	TargetSeat    int    `json:"target_seat,omitempty"`

	// Estado de partida (host -> clientes).
	State                *truco.Snapshot `json:"state,omitempty"`
	FullState            *truco.Snapshot `json:"full_state,omitempty"`
	HostSeat             int             `json:"host_seat"`
	HandoffPort          int             `json:"handoff_port,omitempty"`
	PeerHosts            map[int]string  `json:"peer_hosts,omitempty"`
	SeatSessionIDs       map[int]string  `json:"seat_session_ids,omitempty"`
	TLSSeed              string          `json:"tls_seed,omitempty"`
	Epoch                int             `json:"epoch,omitempty"`
	AuthorityFingerprint string          `json:"authority_fingerprint,omitempty"`
	RouteHint            string          `json:"route_hint,omitempty"`
	RelayHostAdminToken  string          `json:"relay_host_admin_token,omitempty"`

	// Heartbeat opcional para monitorar conectividade.
	HeartbeatUnix int64 `json:"heartbeat_unix,omitempty"`
}

func writeMessage(conn net.Conn, msg Message) error {
	if msg.ProtocolVersion == 0 {
		msg.ProtocolVersion = protocolVersion
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	payload := append(b, '\n')
	if len(payload) > scannerMaxBuffer {
		return errors.New("mensagem excede limite")
	}
	if err := conn.SetWriteDeadline(time.Now().Add(writeMessageDeadline)); err != nil {
		return err
	}
	for len(payload) > 0 {
		n, werr := conn.Write(payload)
		if werr != nil {
			err = werr
			break
		}
		payload = payload[n:]
	}
	_ = conn.SetWriteDeadline(time.Time{})
	return err
}

func newConnReader(conn net.Conn) *bufio.Reader {
	return bufio.NewReaderSize(conn, scannerInitialBuffer)
}

func readMessage(conn net.Conn, reader *bufio.Reader) (Message, error) {
	var msg Message
	if err := conn.SetReadDeadline(time.Now().Add(readMessageDeadline)); err != nil {
		return msg, err
	}
	defer func() {
		_ = conn.SetReadDeadline(time.Time{})
	}()
	line, isPrefix, err := reader.ReadLine()
	if isPrefix {
		// Line exceeds bufio buffer; read additional chunks up to the hard limit.
		full := append([]byte(nil), line...)
		for isPrefix {
			if len(full) > scannerMaxBuffer {
				return msg, errors.New("mensagem excede limite")
			}
			line, isPrefix, err = reader.ReadLine()
			if err != nil {
				break
			}
			full = append(full, line...)
		}
		line = full
	}
	if err != nil {
		if errors.Is(err, io.EOF) && len(line) == 0 {
			return msg, errors.New("conexão encerrada")
		}
		if !errors.Is(err, io.EOF) {
			return msg, err
		}
		if len(line) == 0 {
			return msg, err
		}
	}
	if len(line) > scannerMaxBuffer {
		return msg, errors.New("mensagem excede limite")
	}
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return msg, errors.New("mensagem vazia")
	}
	if err := json.Unmarshal(line, &msg); err != nil {
		return msg, err
	}
	return msg, nil
}

func normalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("nome vazio")
	}
	if len([]rune(name)) > maxPlayerNameLen {
		return "", fmt.Errorf("nome muito longo (máx %d)", maxPlayerNameLen)
	}
	return name, nil
}

func normalizeChatText(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("mensagem vazia")
	}
	if len([]rune(text)) > maxChatTextLen {
		return "", fmt.Errorf("mensagem muito longa (máx %d)", maxChatTextLen)
	}
	return text, nil
}

func normalizeDesiredRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "partner", "opponent", "auto":
		return strings.ToLower(strings.TrimSpace(role))
	default:
		return "auto"
	}
}

func validateGameAction(action string, cardIndex int) error {
	switch action {
	case "play":
		if cardIndex < 0 || cardIndex > 2 {
			return errors.New("índice de carta inválido")
		}
		return nil
	case "truco", "accept", "refuse":
		return nil
	default:
		return errors.New("ação inválida")
	}
}
