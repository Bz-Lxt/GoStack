package control

import (
	"sync"
	"time"
)

type Step struct {
	mu      sync.Mutex
	enabled bool
	pps     int
	last    time.Time
}

func NewStep(pps int) *Step {
	if pps <= 0 {
		pps = 20
	}
	return &Step{pps: pps}
}

func (s *Step) Set(enabled bool, pps int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled
	if pps > 0 {
		s.pps = pps
	}
	if s.pps > 20 {
		s.pps = 20
	}
}

func (s *Step) Snapshot() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{"enabled": s.enabled, "pps": s.pps}
}

func (s *Step) Wait() {
	s.mu.Lock()
	en := s.enabled
	pps := s.pps
	last := s.last
	s.mu.Unlock()
	if !en || pps <= 0 {
		return
	}
	gap := time.Second / time.Duration(pps)
	wait := gap - time.Since(last)
	if wait > 0 {
		time.Sleep(wait)
	}
	s.mu.Lock()
	s.last = time.Now()
	s.mu.Unlock()
}
