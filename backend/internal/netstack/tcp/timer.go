package tcp

import (
	"sync"
	"time"
)

type TimerWheel struct {
	mu    sync.Mutex
	items map[string]*time.Timer
}

func NewTimerWheel() *TimerWheel {
	return &TimerWheel{items: map[string]*time.Timer{}}
}

func (w *TimerWheel) AfterFunc(key string, d time.Duration, fn func()) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if t, ok := w.items[key]; ok {
		t.Stop()
		delete(w.items, key)
	}
	w.items[key] = time.AfterFunc(d, func() {
		w.mu.Lock()
		delete(w.items, key)
		w.mu.Unlock()
		fn()
	})
}

func (w *TimerWheel) Stop(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if t, ok := w.items[key]; ok {
		t.Stop()
		delete(w.items, key)
	}
}

func (w *TimerWheel) StopAll() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for k, t := range w.items {
		t.Stop()
		delete(w.items, k)
	}
}
