package telemetry

import (
	"sync"
	"sync/atomic"
)

type Bus struct {
	mu      sync.Mutex
	buf     []Event
	head    int
	n       int
	cap     int
	dropped atomic.Uint64
	subs    map[int]chan Event
	next    int
}

func NewBus(cap int) *Bus {
	if cap <= 0 {
		cap = 8192
	}
	return &Bus{buf: make([]Event, cap), cap: cap, subs: map[int]chan Event{}}
}

func (b *Bus) Publish(ev Event) {
	b.mu.Lock()
	if b.n == b.cap {
		b.head = (b.head + 1) % b.cap
		b.n--
		b.dropped.Add(1)
	}
	idx := (b.head + b.n) % b.cap
	b.buf[idx] = ev
	b.n++
	subs := make([]chan Event, 0, len(b.subs))
	for _, ch := range b.subs {
		subs = append(subs, ch)
	}
	b.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- ev:
		default:
			b.dropped.Add(1)
		}
	}
}

func (b *Bus) Subscribe(buf int) (int, <-chan Event) {
	if buf <= 0 {
		buf = 256
	}
	ch := make(chan Event, buf)
	b.mu.Lock()
	id := b.next
	b.next++
	b.subs[id] = ch
	b.mu.Unlock()
	return id, ch
}

func (b *Bus) Unsubscribe(id int) {
	b.mu.Lock()
	if ch, ok := b.subs[id]; ok {
		delete(b.subs, id)
		close(ch)
	}
	b.mu.Unlock()
}

func (b *Bus) Dropped() uint64 { return b.dropped.Load() }
