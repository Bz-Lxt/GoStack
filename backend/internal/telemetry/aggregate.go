package telemetry

import (
	"sync"
	"time"

	"gostack/internal/timeutil"
)

type Accum struct {
	mu          sync.Mutex
	goodput     uint64
	throughput  uint64
	retrans     uint64
	last        map[string]map[string]any
}

func NewAccum() *Accum {
	return &Accum{last: map[string]map[string]any{}}
}

func (a *Accum) AddData(n uint64) {
	a.mu.Lock()
	a.goodput += n
	a.throughput += n
	a.mu.Unlock()
}

func (a *Accum) AddWire(n uint64) {
	a.mu.Lock()
	a.throughput += n
	a.mu.Unlock()
}

func (a *Accum) AddRetrans() {
	a.mu.Lock()
	a.retrans++
	a.mu.Unlock()
}

func (a *Accum) Remember(id string, snap map[string]any) {
	a.mu.Lock()
	a.last[id] = snap
	a.mu.Unlock()
}

func (a *Accum) Flush() Event {
	a.mu.Lock()
	defer a.mu.Unlock()
	payload := map[string]any{
		"goodput_bps":    a.goodput * 10,
		"throughput_bps": a.throughput * 10,
		"retransmits":    a.retrans,
		"conns":          len(a.last),
	}
	if len(a.last) == 1 {
		for _, s := range a.last {
			for k, v := range s {
				payload[k] = v
			}
		}
	}
	a.goodput = 0
	a.throughput = 0
	a.retrans = 0
	return Event{
		V: 1, TS: timeutil.FormatRFC3339(timeutil.Now()),
		Type: "aggregate.snapshot", Precise: false, Payload: payload,
	}
}

func Tick(bus *Bus, a *Accum, stop <-chan struct{}) {
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			bus.Publish(a.Flush())
		}
	}
}
