package tcp

import "sync"

type Table struct {
	mu    sync.RWMutex
	conns map[Quad]*Conn
}

func NewTable() *Table {
	return &Table{conns: map[Quad]*Conn{}}
}

func (t *Table) Get(q Quad) *Conn {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.conns[q]
}

func (t *Table) Put(q Quad, c *Conn) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.conns[q] = c
}

func (t *Table) Delete(q Quad) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.conns, q)
}

func (t *Table) List() []*Conn {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]*Conn, 0, len(t.conns))
	for _, c := range t.conns {
		out = append(out, c)
	}
	return out
}

func (t *Table) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.conns)
}
