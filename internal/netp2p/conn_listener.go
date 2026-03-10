package netp2p

import (
	"errors"
	"net"
	"sync"
)

// connListener implementa net.Listener sobre um canal de conexões.
type connListener struct {
	mu     sync.Mutex
	addr   net.Addr
	ch     chan net.Conn
	closed bool
}

func newConnListener(addr net.Addr) *connListener {
	return &connListener{
		addr: addr,
		ch:   make(chan net.Conn, 64),
	}
}

func (l *connListener) push(c net.Conn) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return errors.New("listener fechado")
	}
	l.ch <- c
	return nil
}

func (l *connListener) Accept() (net.Conn, error) {
	c, ok := <-l.ch
	if !ok {
		return nil, errors.New("listener fechado")
	}
	return c, nil
}

func (l *connListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true
	close(l.ch)
	return nil
}

func (l *connListener) Addr() net.Addr {
	return l.addr
}
