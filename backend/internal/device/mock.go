package device

import (
	"context"
	"errors"
	"sync"
)

type Mock struct {
	name string
	mtu  int
	rx   chan []byte
	tx   chan []byte
	mu   sync.Mutex
	dead bool
	ctx  context.Context
	stop context.CancelFunc
}

func NewMock(name string, mtu int) *Mock {
	if mtu <= 0 {
		mtu = 1500
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Mock{
		name: name,
		mtu:  mtu,
		rx:   make(chan []byte, 256),
		tx:   make(chan []byte, 256),
		ctx:  ctx,
		stop: cancel,
	}
}

func Pipe(mtu int) (*Mock, *Mock) {
	a := NewMock("mock-a", mtu)
	b := NewMock("mock-b", mtu)
	a.tx, a.rx = b.rx, b.tx
	return a, b
}

func (m *Mock) Name() string { return m.name }
func (m *Mock) MTU() int     { return m.mtu }

func (m *Mock) Read(p []byte) (int, error) {
	select {
	case <-m.ctx.Done():
		return 0, errors.New("device closed")
	case b := <-m.rx:
		n := copy(p, b)
		return n, nil
	}
}

func (m *Mock) Write(p []byte) (int, error) {
	m.mu.Lock()
	dead := m.dead
	m.mu.Unlock()
	if dead {
		return 0, errors.New("device closed")
	}
	cp := CopyBuf(p)
	select {
	case <-m.ctx.Done():
		return 0, errors.New("device closed")
	case m.tx <- cp:
		return len(p), nil
	}
}

func (m *Mock) Inject(p []byte) {
	select {
	case m.rx <- CopyBuf(p):
	case <-m.ctx.Done():
	default:
	}
}

func (m *Mock) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.dead {
		return nil
	}
	m.dead = true
	m.stop()
	return nil
}
