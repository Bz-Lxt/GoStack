package tcp

import "sync"

const (
	ephemeralMin = 49152
	ephemeralMax = 65535
)

type PortAlloc struct {
	mu   sync.Mutex
	next uint16
	used map[uint16]int
}

func NewPortAlloc() *PortAlloc {
	return &PortAlloc{next: ephemeralMin, used: map[uint16]int{}}
}

func (p *PortAlloc) Reserve(port uint16) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.used[port] > 0 && port != 0 {
		return false
	}
	p.used[port]++
	return true
}

func (p *PortAlloc) Release(port uint16) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.used[port] > 0 {
		p.used[port]--
		if p.used[port] == 0 {
			delete(p.used, port)
		}
	}
}

func (p *PortAlloc) Ephemeral() (uint16, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := 0; i < ephemeralMax-ephemeralMin+1; i++ {
		port := p.next
		p.next++
		if p.next < ephemeralMin || p.next > ephemeralMax {
			p.next = ephemeralMin
		}
		if p.used[port] == 0 {
			p.used[port]++
			return port, true
		}
	}
	return 0, false
}
