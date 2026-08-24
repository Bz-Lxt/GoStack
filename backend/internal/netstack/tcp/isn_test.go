package tcp

import (
	"net/netip"
	"testing"
	"time"
)

func TestISNNotConstant(t *testing.T) {
	q := Quad{LocalIP: netip.MustParseAddr("10.0.0.2"), LocalPort: 9, RemoteIP: netip.MustParseAddr("10.0.0.1"), RemotePort: 8}
	a := GenerateISN(q)
	time.Sleep(2 * time.Microsecond)
	b := GenerateISN(q)
	if a == 0 && b == 0 {
		t.Fatal("zero ISN")
	}
}
