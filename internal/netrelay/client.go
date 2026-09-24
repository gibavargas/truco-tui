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

	helloWriteTimeout = 3 * time.Second
	ackReadTimeout    = 5 * time.Second
	tcpDialTimeout    = 5 * time.Second
	tcpKeepAlive      = 15 * time.Second
)

type RelayHTTPError struct {
	Status  int
	Code    string
	Message string
}

func (e RelayHTTPError) Error() string {
	if strings.TrimSpace(e.Code) != "" && strings.TrimSpace(e.Message) != "" {
		return fmt.Sprintf("relay http status %d (%s): %s", e.Status, e.Code, e.Message)
	}
	if strings.TrimSpace(e.Message) != "" {
		return fmt.Sprintf("relay http status %d: %s", e.Status, e.Message)
	}
	if strings.TrimSpace(e.Code) != "" {
		return fmt.Sprintf("relay http status %d (%s)", e.Status, e.Code)
	}
	return fmt.Sprintf("relay http status %d", e.Status)
}

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
	return postJSON(relayURL, sec, "/v2/heartbeat", req, &out)
}

// OpenPeerTunnel conecta ao relay (QUIC primeiro, TCP/TLS como fallback) e
// abre um túnel bidirecional até a autoridade da sessão.
func OpenPeerTunnel(ctx context.Context, sec ClientSecurity, quicAddr, tcpAddr, sessionID, peerID, credential, targetPeerID string) (net.Conn, error) {
	hello := tunnelHello{
		Type:         "peer_tunnel",
		SessionID:    sessionID,
		PeerID:       peerID,
		Credential:   credential,
		TargetPeerID: targetPeerID,
	}
	return openRelayStream(ctx, sec, quicAddr, tcpAddr, hello, true)
}

// openRelayStream tenta QUIC e, em falha de transporte, tenta TCP/TLS.
// closeQuicConn indica se a conexão QUIC deve ser fechada junto do stream.
func openRelayStream(ctx context.Context, sec ClientSecurity, quicAddr, tcpAddr string, hello tunnelHello, closeQuicConn bool) (net.Conn, error) {
	quicAddr = strings.TrimSpace(quicAddr)
	tcpAddr = strings.TrimSpace(tcpAddr)
	var quicErr error
	if quicAddr != "" {
		qctx, cancel := quicDialBudget(ctx)
		conn, _, err := openRelayQUICStream(qctx, sec, quicAddr, hello, closeQuicConn)
		cancel()
		if err == nil {
			return conn, nil
		}
		quicErr = err
	}
	if tcpAddr != "" {
		conn, _, err := dialRelayTCPConn(ctx, sec, tcpAddr, hello)
		return conn, err
	}
	if quicErr != nil {
		return nil, quicErr
	}
	return nil, errors.New("relay addr ausente")
}

// quicDialBudget reserva parte do deadline do contexto para a tentativa QUIC,
// deixando o restante para o fallback TCP.
func quicDialBudget(ctx context.Context) (context.Context, context.CancelFunc) {
	if dl, ok := ctx.Deadline(); ok {
		quic := time.Until(dl) * 3 / 5
		if quic < 500*time.Millisecond {
			quic = 500 * time.Millisecond
		}
		return context.WithTimeout(ctx, quic)
	}
	return context.WithTimeout(ctx, 3*time.Second)
}

func openRelayQUICStream(ctx context.Context, sec ClientSecurity, quicAddr string, hello tunnelHello, closeConn bool) (net.Conn, quic.Connection, error) {
	conn, err := quic.DialAddr(ctx, quicAddr, relayTLSConfig(sec), relayQUICConfig())
	if err != nil {
		return nil, nil, err
	}
	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		_ = conn.CloseWithError(0, "open stream failed")
		return nil, nil, err
	}
	if err := writeHello(stream, hello); err != nil {
		_ = stream.Close()
		_ = conn.CloseWithError(0, "hello failed")
		return nil, nil, err
	}
	return netquic.NewStreamConn(conn, stream, closeConn), conn, nil
}

// dialRelayTCPConn abre uma conexão TCP+TLS com o relay e escreve o hello.
// Retorna também o bufio.Reader para não descartar bytes já bufferizados.
func dialRelayTCPConn(ctx context.Context, sec ClientSecurity, tcpAddr string, hello tunnelHello) (net.Conn, *bufio.Reader, error) {
	d := &net.Dialer{
		Timeout:   tcpDialTimeout,
		KeepAlive: tcpKeepAlive,
	}
	raw, err := d.DialContext(ctx, "tcp", tcpAddr)
	if err != nil {
		return nil, nil, err
	}
	tconn := tls.Client(raw, relayTLSConfig(sec))
	if err := tconn.HandshakeContext(ctx); err != nil {
		_ = raw.Close()
		return nil, nil, err
	}
	if err := writeHello(tconn, hello); err != nil {
		_ = tconn.Close()
		return nil, nil, err
	}
	return tconn, bufio.NewReader(tconn), nil
}

type HostAcceptor struct {
	conn       quic.Connection
	signal     net.Conn
	signalIn   *bufio.Reader
	tcpAddr    string
	sec        ClientSecurity
	sessionID  string
	peerID     string
	credential string
	accept     chan net.Conn
	errc       chan error
	closed     chan struct{}
}

func OpenHostAcceptor(ctx context.Context, sec ClientSecurity, quicAddr, tcpAddr, sessionID, peerID, credential string) (*HostAcceptor, error) {
	quicAddr = strings.TrimSpace(quicAddr)
	tcpAddr = strings.TrimSpace(tcpAddr)
	hello := tunnelHello{
		Type:       "host_register",
		SessionID:  sessionID,
		PeerID:     peerID,
		Credential: credential,
	}
	if quicAddr != "" {
		qctx, cancel := quicDialBudget(ctx)
		conn, qconn, err := openRelayQUICStream(qctx, sec, quicAddr, hello, false)
		if err == nil {
			denied, aerr := readRegisterAck(conn)
			_ = conn.Close() // fecha apenas o stream de registro; a conexão segue
			if aerr == nil {
				cancel()
				if denied {
					_ = qconn.CloseWithError(0, "register denied")
					return nil, errors.New("registro no relay rejeitado")
				}
				a := &HostAcceptor{
					conn:       qconn,
					accept:     make(chan net.Conn, 32),
					errc:       make(chan error, 1),
					closed:     make(chan struct{}),
				}
				go a.acceptLoop()
				return a, nil
			}
			_ = qconn.CloseWithError(0, "register ack failed")
		}
		cancel()
	}
	if tcpAddr != "" {
		conn, reader, err := dialRelayTCPConn(ctx, sec, tcpAddr, hello)
		if err != nil {
			return nil, err
		}
		denied, aerr := readRegisterAckBuffered(conn, reader)
		if aerr != nil {
			_ = conn.Close()
			return nil, aerr
		}
		if denied {
			_ = conn.Close()
			return nil, errors.New("registro no relay rejeitado")
		}
		a := &HostAcceptor{
			signal:     conn,
			signalIn:   reader,
			tcpAddr:    tcpAddr,
			sec:        sec,
			sessionID:  sessionID,
			peerID:     peerID,
			credential: credential,
			accept:     make(chan net.Conn, 32),
			errc:       make(chan error, 1),
			closed:     make(chan struct{}),
		}
		go a.acceptLoop()
		return a, nil
	}
	return nil, errors.New("relay addr ausente")
}

// readRegisterAck lê a resposta do host_register em modo QUIC.
func readRegisterAck(conn net.Conn) (denied bool, err error) {
	return readRegisterAckBuffered(conn, bufio.NewReader(conn))
}

func readRegisterAckBuffered(conn net.Conn, reader *bufio.Reader) (denied bool, err error) {
	if err := conn.SetReadDeadline(time.Now().Add(ackReadTimeout)); err != nil {
		return false, err
	}
	defer func() {
		_ = conn.SetReadDeadline(time.Time{})
	}()
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return false, err
	}
	var ack struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(line), &ack); err != nil {
		return false, err
	}
	if !ack.OK {
		if strings.TrimSpace(ack.Error) != "" {
			return true, errors.New(ack.Error)
		}
		return true, nil
	}
	return false, nil
}

func (a *HostAcceptor) acceptLoop() {
	defer close(a.accept)
	defer close(a.errc)
	if a.conn != nil {
		a.acceptLoopQUIC()
		return
	}
	a.acceptLoopTCP()
}

func (a *HostAcceptor) acceptLoopQUIC() {
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

// acceptLoopTCP mantém a conexão de sinal e disca uma conexão de dados para
// cada tunnel_open anunciado pelo relay.
func (a *HostAcceptor) acceptLoopTCP() {
	for {
		line, err := a.signalIn.ReadBytes('\n')
		if err != nil {
			select {
			case <-a.closed:
			default:
				a.errc <- err
			}
			return
		}
		var msg struct {
			Type     string `json:"type"`
			TunnelID string `json:"tunnel_id"`
		}
		if json.Unmarshal(bytes.TrimSpace(line), &msg) != nil || msg.Type != "tunnel_open" {
			continue
		}
		hello := tunnelHello{
			Type:       "tunnel_accept",
			SessionID:  a.sessionID,
			PeerID:     a.peerID,
			Credential: a.credential,
			TunnelID:   msg.TunnelID,
		}
		dctx, dcancel := context.WithTimeout(context.Background(), tcpDialTimeout)
		conn, _, derr := dialRelayTCPConn(dctx, a.sec, a.tcpAddr, hello)
		dcancel()
		if derr != nil {
			continue
		}
		select {
		case <-a.closed:
			_ = conn.Close()
			return
		case a.accept <- conn:
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
	if a.signal != nil {
		return a.signal.Close()
	}
	if a.conn != nil {
		return a.conn.CloseWithError(0, "host acceptor closed")
	}
	return nil
}

func (a *HostAcceptor) Addr() net.Addr {
	if a.signal != nil {
		return a.signal.LocalAddr()
	}
	if a.conn == nil {
		return nil
	}
	return a.conn.LocalAddr()
}

type deadlineWriter interface {
	Write(p []byte) (int, error)
	SetWriteDeadline(t time.Time) error
}

func writeHello(w deadlineWriter, hello tunnelHello) error {
	b, err := json.Marshal(hello)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := w.SetWriteDeadline(time.Now().Add(helloWriteTimeout)); err != nil {
		return err
	}
	defer func() {
		_ = w.SetWriteDeadline(time.Time{})
	}()
	_, err = w.Write(b)
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
			Error     string `json:"error"`
			ErrorCode string `json:"error_code"`
		}
		if derr := json.NewDecoder(resp.Body).Decode(&apiErr); derr == nil && strings.TrimSpace(apiErr.Error) != "" {
			return RelayHTTPError{
				Status:  resp.StatusCode,
				Code:    strings.TrimSpace(apiErr.ErrorCode),
				Message: strings.TrimSpace(apiErr.Error),
			}
		}
		return RelayHTTPError{Status: resp.StatusCode}
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
