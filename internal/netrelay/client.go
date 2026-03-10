package netrelay

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/quic-go/quic-go"

	"truco-tui/internal/netquic"
)

const (
	httpTimeout = 5 * time.Second
)

type ClientSecurity struct {
	RelaySPKIPin string
	RootCAs      *x509.CertPool
	ServerName   string
}

func CreateSession(relayURL string, sec ClientSecurity, req CreateSessionRequest) (CreateSessionResponse, error) {
	var out CreateSessionResponse
	if err := postJSON(relayURL, sec, "/v2/create-session", req, &out); err != nil {
		return out, err
	}
	return out, nil
}

func MintJoinTicket(relayURL string, sec ClientSecurity, req MintJoinTicketRequest) (MintJoinTicketResponse, error) {
	var out MintJoinTicketResponse
	if err := postJSON(relayURL, sec, "/v2/mint-join-ticket", req, &out); err != nil {
		return out, err
	}
	return out, nil
}

func JoinSession(relayURL string, sec ClientSecurity, req JoinSessionRequest) (JoinSessionResponse, error) {
	var out JoinSessionResponse
	if err := postJSON(relayURL, sec, "/v2/join-session", req, &out); err != nil {
		return out, err
	}
	return out, nil
}

func PublishAuthority(relayURL string, sec ClientSecurity, req PublishAuthorityRequest) (PublishAuthorityResponse, error) {
	var out PublishAuthorityResponse
	if err := postJSON(relayURL, sec, "/v2/publish-authority", req, &out); err != nil {
		return out, err
	}
	return out, nil
}

func Heartbeat(relayURL string, sec ClientSecurity, req HeartbeatRequest) error {
	var out map[string]any
	return postJSON(relayURL, sec, "/v1/heartbeat", req, &out)
}

func OpenPeerTunnel(ctx context.Context, sec ClientSecurity, quicAddr, sessionID, peerID, credential, targetPeerID string) (net.Conn, error) {
	if strings.TrimSpace(quicAddr) == "" {
		return nil, errors.New("quic relay addr ausente")
	}
	conn, err := quic.DialAddr(ctx, quicAddr, relayTLSConfig(sec), relayQUICConfig())
	if err != nil {
		return nil, err
	}
	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		_ = conn.CloseWithError(0, "open stream failed")
		return nil, err
	}
	hello := tunnelHello{
		Type:         "peer_tunnel",
		SessionID:    sessionID,
		PeerID:       peerID,
		Credential:   credential,
		TargetPeerID: targetPeerID,
	}
	if err := writeHello(stream, hello); err != nil {
		_ = stream.Close()
		_ = conn.CloseWithError(0, "hello failed")
		return nil, err
	}
	return netquic.NewStreamConn(conn, stream, true), nil
}

type HostAcceptor struct {
	conn   quic.Connection
	accept chan net.Conn
	errc   chan error
	closed chan struct{}
}

func OpenHostAcceptor(ctx context.Context, sec ClientSecurity, quicAddr, sessionID, peerID, credential string) (*HostAcceptor, error) {
	if strings.TrimSpace(quicAddr) == "" {
		return nil, errors.New("quic relay addr ausente")
	}
	conn, err := quic.DialAddr(ctx, quicAddr, relayTLSConfig(sec), relayQUICConfig())
	if err != nil {
		return nil, err
	}
	registerStream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		_ = conn.CloseWithError(0, "register stream failed")
		return nil, err
	}
	hello := tunnelHello{
		Type:       "host_register",
		SessionID:  sessionID,
		PeerID:     peerID,
		Credential: credential,
	}
	if err := writeHello(registerStream, hello); err != nil {
		_ = registerStream.Close()
		_ = conn.CloseWithError(0, "register failed")
		return nil, err
	}
	ackReader := bufio.NewReader(registerStream)
	line, err := ackReader.ReadBytes('\n')
	if err != nil {
		_ = registerStream.Close()
		_ = conn.CloseWithError(0, "register ack failed")
		return nil, err
	}
	var ack map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(line), &ack); err != nil {
		_ = registerStream.Close()
		_ = conn.CloseWithError(0, "invalid register ack")
		return nil, err
	}
	if ok, _ := ack["ok"].(bool); !ok {
		_ = registerStream.Close()
		_ = conn.CloseWithError(0, "register denied")
		return nil, errors.New("registro no relay rejeitado")
	}
	_ = registerStream.Close()

	a := &HostAcceptor{
		conn:   conn,
		accept: make(chan net.Conn, 32),
		errc:   make(chan error, 1),
		closed: make(chan struct{}),
	}
	go a.acceptLoop()
	return a, nil
}

func (a *HostAcceptor) acceptLoop() {
	defer close(a.accept)
	defer close(a.errc)
	for {
		stream, err := a.conn.AcceptStream(context.Background())
		if err != nil {
			select {
			case <-a.closed:
			default:
				a.errc <- err
			}
			return
		}
		select {
		case <-a.closed:
			_ = stream.Close()
			return
		case a.accept <- netquic.NewStreamConn(a.conn, stream, false):
		}
	}
}

func (a *HostAcceptor) Accept() (net.Conn, error) {
	select {
	case c, ok := <-a.accept:
		if !ok {
			select {
			case err := <-a.errc:
				if err != nil {
					return nil, err
				}
			default:
			}
			return nil, errors.New("relay acceptor encerrado")
		}
		return c, nil
	case err := <-a.errc:
		if err == nil {
			return nil, errors.New("relay acceptor encerrado")
		}
		return nil, err
	}
}

func (a *HostAcceptor) Close() error {
	select {
	case <-a.closed:
	default:
		close(a.closed)
	}
	if a.conn != nil {
		return a.conn.CloseWithError(0, "host acceptor closed")
	}
	return nil
}

func (a *HostAcceptor) Addr() net.Addr {
	if a.conn == nil {
		return nil
	}
	return a.conn.LocalAddr()
}

func writeHello(stream quic.Stream, hello tunnelHello) error {
	b, err := json.Marshal(hello)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := stream.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return err
	}
	defer stream.SetWriteDeadline(time.Time{})
	_, err = stream.Write(b)
	return err
}

func relayTLSConfig(sec ClientSecurity) *tls.Config {
	cfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
		RootCAs:    sec.RootCAs,
		ServerName: strings.TrimSpace(sec.ServerName),
		NextProtos: []string{TunnelProto},
	}
	if strings.TrimSpace(sec.RelaySPKIPin) != "" {
		cfg.VerifyConnection = verifySPKIPin(strings.TrimSpace(sec.RelaySPKIPin))
	}
	return cfg
}

func relayQUICConfig() *quic.Config {
	return &quic.Config{
		HandshakeIdleTimeout: 5 * time.Second,
		MaxIdleTimeout:       30 * time.Second,
		KeepAlivePeriod:      10 * time.Second,
	}
}

func postJSON(relayURL string, sec ClientSecurity, path string, req any, out any) error {
	base, err := url.Parse(strings.TrimSpace(relayURL))
	if err != nil {
		return err
	}
	if base.Scheme == "" {
		base.Scheme = "https"
	}
	base.Path = strings.TrimRight(base.Path, "/") + path

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
		RootCAs:    sec.RootCAs,
		ServerName: strings.TrimSpace(sec.ServerName),
	}
	if strings.TrimSpace(sec.RelaySPKIPin) != "" {
		tlsCfg.VerifyConnection = verifySPKIPin(strings.TrimSpace(sec.RelaySPKIPin))
	}
	httpClient := &http.Client{
		Timeout: httpTimeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
	}
	reqHTTP, err := http.NewRequest(http.MethodPost, base.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	reqHTTP.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(reqHTTP)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr struct {
			Error string `json:"error"`
		}
		if derr := json.NewDecoder(resp.Body).Decode(&apiErr); derr == nil && strings.TrimSpace(apiErr.Error) != "" {
			return fmt.Errorf("relay http status %d: %s", resp.StatusCode, apiErr.Error)
		}
		return fmt.Errorf("relay http status %d", resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func verifySPKIPin(pin string) func(cs tls.ConnectionState) error {
	pin = strings.TrimSpace(pin)
	return func(cs tls.ConnectionState) error {
		if len(cs.PeerCertificates) == 0 {
			return errors.New("relay certificate missing")
		}
		leaf := cs.PeerCertificates[0]
		sum := sha256.Sum256(leaf.RawSubjectPublicKeyInfo)
		gotHex := hex.EncodeToString(sum[:])
		gotB64 := base64.StdEncoding.EncodeToString(sum[:])
		if pin == gotHex || pin == gotB64 {
			return nil
		}
		return errors.New("relay SPKI pin mismatch")
	}
}
