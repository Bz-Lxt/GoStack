package tcp

import (
	"testing"
	"time"
)

func TestTimerWheel(t *testing.T) {
	w := NewTimerWheel()
	done := make(chan struct{}, 1)
	w.AfterFunc("a", 5*time.Millisecond, func() { done <- struct{}{} })
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timer")
	}
	w.AfterFunc("b", time.Second, func() {})
	w.Stop("b")
	w.StopAll()
}
