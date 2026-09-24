// Command e2e-tcp-smoke valida o data plane TCP do binário truco-relay real:
// host_register (sinal), peer_tunnel, tunnel_open→tunnel_accept (dial-back do
// host) e bridge de bytes — tudo sobre TCP/TLS, simulando UDP bloqueado.
package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

type hello struct {
	Type         string `json:"type"`
	SessionID    string `json:"session_id"`
	PeerID       string `json:"peer_id"`
	Credential   string `json:"credential"`
	TargetPeerID string `json:"target_peer_id,omitempty"`
	TunnelID     string `json:"tunnel_id,omitempty"`
}

func apiClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func apiPost(relayURL, path string, body map[string]any) map[string]any {
	b, _ := json.Marshal(body)
	resp, err := apiClient().Post(relayURL+path, "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Println("api", path, "erro:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		fmt.Println("api", path, "json:", err, string(raw))
		os.Exit(1)
	}
	if resp.StatusCode != 200 {
		fmt.Println("api", path, "status:", resp.StatusCode, string(raw))
		os.Exit(1)
	}
	return m
}

func dialTCP(addr string) net.Conn {
	d := &net.Dialer{Timeout: 5 * time.Second}
	raw, err := d.Dial("tcp", addr)
	if err != nil {
		fmt.Println("dial:", err)
		os.Exit(1)
	}
	c := tls.Client(raw, &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"truco-relay-quic-v2"},
	})
	if err := c.Handshake(); err != nil {
		fmt.Println("tls:", err)
		os.Exit(1)
	}
	return c
}

func writeJSONLine(c net.Conn, v any) {
	b, _ := json.Marshal(v)
	b = append(b, '\n')
	if _, err := c.Write(b); err != nil {
		fmt.Println("write:", err)
		os.Exit(1)
	}
}

func readJSONLine(r *bufio.Reader) map[string]any {
	line, err := r.ReadBytes('\n')
	if err != nil {
		fmt.Println("read:", err)
		os.Exit(1)
	}
	var m map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(line), &m); err != nil {
		fmt.Println("json:", err, string(line))
		os.Exit(1)
	}
	return m
}

func mustGet(m map[string]any, k string) string {
	s, _ := m[k].(string)
	return s
}

func main() {
	relayURL := os.Getenv("RELAY_URL")
	if relayURL == "" {
		relayURL = "https://127.0.0.1:19443"
	}
	tcpAddr := os.Getenv("RELAY_TCP_ADDR")
	if tcpAddr == "" {
		tcpAddr = "127.0.0.1:19445"
	}

	// 1. Control plane: cria sessão via API https do binário.
	fmt.Println("== control plane ==")
	created := apiPost(relayURL, "/v2/create-session", map[string]any{"host_identity": "E2EHost", "num_players": 2})
	sessID := mustGet(created, "session_id")
	adminTok := mustGet(created, "host_admin_token")
	hostPeer := mustGet(created, "host_peer_id")
	hostCred := mustGet(created, "host_peer_credential")
	fmt.Println("sessão:", sessID)
	if mustGet(created, "tcp_addr") == "" {
		fmt.Println("FALHOU: resposta sem tcp_addr")
		os.Exit(1)
	}

	// 2. Host registra via TCP (QUIC "bloqueado" — nem tentamos).
	fmt.Println("== host_register via TCP ==")
	signal := dialTCP(tcpAddr)
	defer signal.Close()
	writeJSONLine(signal, hello{Type: "host_register", SessionID: sessID, PeerID: hostPeer, Credential: hostCred})
	signalIn := bufio.NewReader(signal)
	if err := signal.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ack := readJSONLine(signalIn)
	if ok, _ := ack["ok"].(bool); !ok {
		fmt.Println("FALHOU: register ack:", ack)
		os.Exit(1)
	}
	_ = signal.SetReadDeadline(time.Time{})
	fmt.Println("host registrado via TCP")

	// 3. Par entra pela API https e abre peer_tunnel via TCP.
	fmt.Println("== peer_tunnel via TCP ==")
	minted := apiPost(relayURL, "/v2/mint-join-ticket", map[string]any{
		"session_id": sessID, "host_admin_token": adminTok, "player_name": "E2EGuest",
	})
	joined := apiPost(relayURL, "/v2/join-session", map[string]any{
		"session_id": sessID, "join_ticket": mustGet(minted, "join_ticket"), "player_name": "E2EGuest",
	})
	peerID := mustGet(joined, "peer_id")
	peerCred := mustGet(joined, "peer_credential")
	authority := mustGet(joined, "authority_peer_id")

	peer := dialTCP(tcpAddr)
	defer peer.Close()
	writeJSONLine(peer, hello{Type: "peer_tunnel", SessionID: sessID, PeerID: peerID, Credential: peerCred, TargetPeerID: authority})

	// 4. Host recebe tunnel_open e disca conexão de dados (tunnel_accept).
	if err := signal.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	open := readJSONLine(signalIn)
	_ = signal.SetReadDeadline(time.Time{})
	if mustGet(open, "type") != "tunnel_open" {
		fmt.Println("FALHOU: esperado tunnel_open, veio:", open)
		os.Exit(1)
	}
	tunnelID := mustGet(open, "tunnel_id")
	fmt.Println("tunnel_open recebido:", tunnelID)

	data := dialTCP(tcpAddr)
	defer data.Close()
	writeJSONLine(data, hello{Type: "tunnel_accept", SessionID: sessID, PeerID: hostPeer, Credential: hostCred, TunnelID: tunnelID})

	// 5. Bridge: ping do par deve chegar ao host e voltar como pong.
	if _, err := peer.Write([]byte("ping")); err != nil {
		fmt.Println("peer write:", err)
		os.Exit(1)
	}
	if err := data.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	buf := make([]byte, 4)
	if _, err := io.ReadFull(data, buf); err != nil {
		fmt.Println("host read:", err)
		os.Exit(1)
	}
	if string(buf) != "ping" {
		fmt.Printf("FALHOU: host recebeu %q\n", string(buf))
		os.Exit(1)
	}
	if _, err := data.Write([]byte("pong")); err != nil {
		fmt.Println("host write:", err)
		os.Exit(1)
	}
	if err := peer.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if _, err := io.ReadFull(peer, buf); err != nil {
		fmt.Println("peer read:", err)
		os.Exit(1)
	}
	if string(buf) != "pong" {
		fmt.Printf("FALHOU: peer recebeu %q\n", string(buf))
		os.Exit(1)
	}
	fmt.Println("ping/pong através do túnel TCP: OK")
	fmt.Println("E2E OK — data plane TCP do binário relay funcional com UDP bloqueado")
}
