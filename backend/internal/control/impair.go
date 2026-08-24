package control

import (
	"math/rand"
	"sync"
	"time"
)

type Impair struct {
	mu       sync.Mutex
	Loss     float64
	Delay    time.Duration
	Jitter   time.Duration
	Reorder  float64
	hold     [][]byte
	out      chan []byte
	once     sync.Once
}

func NewImpair() *Impair {
	return &Impair{out: make(chan []byte, 256)}
}

func (i *Impair) Configure(loss float64, delay, jitter time.Duration, reorder float64) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Loss = loss
	i.Delay = delay
	i.Jitter = jitter
	i.Reorder = reorder
}

func (i *Impair) Snapshot() map[string]any {
	i.mu.Lock()
	defer i.mu.Unlock()
	return map[string]any{
		"loss": i.Loss, "delay_ms": i.Delay.Milliseconds(),
		"jitter_ms": i.Jitter.Milliseconds(), "reorder": i.Reorder,
	}
}

func (i *Impair) Drop() bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.Loss <= 0 {
		return false
	}
	return rand.Float64() < i.Loss
}

func (i *Impair) MaybeDelay(pkt []byte) []byte {
	i.mu.Lock()
	d, j, r := i.Delay, i.Jitter, i.Reorder
	i.mu.Unlock()
	if r > 0 && rand.Float64() < r {
		i.mu.Lock()
		i.hold = append(i.hold, append([]byte(nil), pkt...))
		if len(i.hold) >= 2 {
			out := i.hold[0]
			i.hold = i.hold[1:]
			i.mu.Unlock()
			return out
		}
		i.mu.Unlock()
		return nil
	}
	if d <= 0 && j <= 0 {
		return pkt
	}
	wait := d
	if j > 0 {
		wait += time.Duration(rand.Int63n(int64(j) + 1))
	}
	cp := append([]byte(nil), pkt...)
	time.Sleep(wait)
	return cp
}
