package tcp

import (
	"net"
	"sync"
	"time"
)

type Listener struct {
	mu       sync.Mutex
	host     Host
	ip       Addr
	backlog  chan *Conn
	closed   bool
	dead     chan struct{}
	once     sync.Once
	acceptDL time.Time
}

func NewListener(host Host, ip Addr, backlog int) *Listener {
	if backlog <= 0 {
		backlog = 128
	}
	return &Listener{
		host:    host,
		ip:      ip,
		backlog: make(chan *Conn, backlog),
		dead:    make(chan struct{}),
	}
}

func (l *Listener) Accept() (net.Conn, error) {
	if !l.acceptDL.IsZero() && !time.Now().Before(l.acceptDL) {
		return nil, ErrTimeout
	}
	var timeout <-chan time.Time
	if !l.acceptDL.IsZero() {
		t := time.NewTimer(time.Until(l.acceptDL))
		defer t.Stop()
		timeout = t.C
	}
	select {
	case c := <-l.backlog:
		return c, nil
	case <-l.dead:
		return nil, ErrClosed
	case <-timeout:
		return nil, ErrTimeout
	}
}

func (l *Listener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true
	l.once.Do(func() { close(l.dead) })
	return nil
}

func (l *Listener) Addr() net.Addr { return l.ip }

func (l *Listener) Closed() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.closed
}

func (l *Listener) Enqueue(c *Conn) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return false
	}
	select {
	case l.backlog <- c:
		return true
	default:
		return false
	}
}

func (l *Listener) Port() uint16 { return l.ip.Port }
