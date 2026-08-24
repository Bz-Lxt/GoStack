package tcp

import (
	"fmt"
	"net/netip"
)

type Quad struct {
	LocalIP    netip.Addr
	LocalPort  uint16
	RemoteIP   netip.Addr
	RemotePort uint16
}

func (q Quad) String() string {
	return fmt.Sprintf("%s:%d-%s:%d", q.RemoteIP, q.RemotePort, q.LocalIP, q.LocalPort)
}

func (q Quad) Reverse() Quad {
	return Quad{
		LocalIP: q.RemoteIP, LocalPort: q.RemotePort,
		RemoteIP: q.LocalIP, RemotePort: q.LocalPort,
	}
}

func (q Quad) Local() string  { return fmt.Sprintf("%s:%d", q.LocalIP, q.LocalPort) }
func (q Quad) Remote() string { return fmt.Sprintf("%s:%d", q.RemoteIP, q.RemotePort) }
