package netp2p

import (
	"bytes"
	"context"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestClientCloseWithNilConnDoesNotPanic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &ClientSession{
		ctx:    ctx,
		cancel: cancel,
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("second Close() error = %v, want nil", err)
	}
}

func TestClientSendWithoutConnectionReturnsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := &ClientSession{
		ctx:    ctx,
		cancel: cancel,
	}
	err := c.send(Message{Type: "ping"})
	if err == nil {
		t.Fatalf("send() error = nil, want error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "conex") {
		t.Fatalf("send() error = %q, want connection-related error", err.Error())
	}
}

func TestReadMessageKeepsDeadlineWhileReadingLongLine(t *testing.T) {
	conn := &deadlineProbeConn{
		payload: bytes.Repeat([]byte("a"), scannerInitialBuffer+8),
	}
	reader := newConnReader(conn)
	_, _ = readMessage(conn, reader)
	if conn.readCalls < 2 {
		t.Fatalf("expected at least 2 reads, got %d", conn.readCalls)
	}
	if !conn.secondReadHadDeadline {
		t.Fatalf("expected deadline to remain active for continuation reads")
	}
}

func TestReplacementInviteRemainsValidAfterJoinAckFailure(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 2)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	guest, err := JoinSession(key, "Guest", "auto")
	if err != nil {
		t.Fatalf("JoinSession guest: %v", err)
	}
	defer guest.Close()

	waitUntil(t, 2*time.Second, host.IsFull, "host lobby full")
	if err := host.StartGame(); err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	_ = guest.Close()
	waitUntil(t, 2*time.Second, func() bool { return !host.IsSeatConnected(1) }, "seat 2 should disconnect")

	replKey, err := host.RequestReplacementInvite(0, 1)
	if err != nil {
		t.Fatalf("RequestReplacementInvite: %v", err)
	}
	inv, err := DecodeInviteKey(replKey)
	if err != nil {
		t.Fatalf("DecodeInviteKey(replacement): %v", err)
	}

	serverConn, clientConn := net.Pipe()
	done := make(chan struct{})
	go func() {
		host.handleConn(&failWriteConn{Conn: serverConn})
		close(done)
	}()

	if err := writeMessage(clientConn, Message{
		Type:            "join",
		ProtocolVersion: protocolVersion,
		Token:           inv.Token,
		Name:            "FlakySub",
		DesiredRole:     "auto",
		ReplaceToken:    inv.ReplaceToken,
	}); err != nil {
		t.Fatalf("writeMessage(join replacement): %v", err)
	}
	closeConnWithLog(clientConn, "test replacement ack failure")

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting failed replacement join")
	}

	sub, err := JoinSession(replKey, "Sub", "auto")
	if err != nil {
		t.Fatalf("replacement invite should remain valid after failed handshake: %v", err)
	}
	defer sub.Close()
	if got := sub.AssignedSeat(); got != 1 {
		t.Fatalf("replacement seat = %d, want 1", got)
	}
}

type failWriteConn struct {
	net.Conn
}

func (c *failWriteConn) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}

type deadlineProbeConn struct {
	payload               []byte
	readCalls             int
	readDeadline          time.Time
	secondReadHadDeadline bool
}

func (c *deadlineProbeConn) Read(p []byte) (int, error) {
	switch c.readCalls {
	case 0:
		c.readCalls++
		n := copy(p, c.payload[:scannerInitialBuffer])
		c.payload = c.payload[n:]
		return n, nil
	case 1:
		c.readCalls++
		c.secondReadHadDeadline = !c.readDeadline.IsZero()
		return 0, io.EOF
	default:
		c.readCalls++
		return 0, io.EOF
	}
}

func (c *deadlineProbeConn) Write(p []byte) (int, error) { return len(p), nil }
func (c *deadlineProbeConn) Close() error                { return nil }
func (c *deadlineProbeConn) LocalAddr() net.Addr         { return staticAddr("local") }
func (c *deadlineProbeConn) RemoteAddr() net.Addr        { return staticAddr("remote") }
func (c *deadlineProbeConn) SetDeadline(t time.Time) error {
	c.readDeadline = t
	return nil
}
func (c *deadlineProbeConn) SetReadDeadline(t time.Time) error {
	c.readDeadline = t
	return nil
}
func (c *deadlineProbeConn) SetWriteDeadline(time.Time) error { return nil }

type staticAddr string

func (a staticAddr) Network() string { return "probe" }
func (a staticAddr) String() string  { return string(a) }
