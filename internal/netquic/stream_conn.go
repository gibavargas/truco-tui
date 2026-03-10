package netquic

import (
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

// StreamConn adapta quic.Stream para net.Conn e opcionalmente fecha a conexão QUIC no Close.
type StreamConn struct {
	conn      quic.Connection
	stream    quic.Stream
	closeConn bool
	once      sync.Once
}

func NewStreamConn(conn quic.Connection, stream quic.Stream, closeConn bool) *StreamConn {
	return &StreamConn{
		conn:      conn,
		stream:    stream,
		closeConn: closeConn,
	}
}

func (c *StreamConn) Read(b []byte) (int, error)  { return c.stream.Read(b) }
func (c *StreamConn) Write(b []byte) (int, error) { return c.stream.Write(b) }

func (c *StreamConn) Close() error {
	var retErr error
	c.once.Do(func() {
		_ = c.stream.Close()
		if c.closeConn && c.conn != nil {
			retErr = c.conn.CloseWithError(0, "closed")
		}
	})
	return retErr
}

func (c *StreamConn) LocalAddr() net.Addr {
	if c.conn == nil {
		return nil
	}
	return c.conn.LocalAddr()
}

func (c *StreamConn) RemoteAddr() net.Addr {
	if c.conn == nil {
		return nil
	}
	return c.conn.RemoteAddr()
}

func (c *StreamConn) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

func (c *StreamConn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

func (c *StreamConn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}
