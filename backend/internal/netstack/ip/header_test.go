package ip_test

import (
	"net/netip"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"gostack/internal/netstack/ip"
)

func TestIPv4RoundTripAndGopacket(t *testing.T) {
	src := netip.MustParseAddr("10.0.0.2")
	dst := netip.MustParseAddr("10.0.0.1")
	payload := []byte{1, 2, 3, 4, 5}
	pkt := ip.Marshal(src, dst, ip.ProtoTCP, 64, 7, payload)
	h, err := ip.Parse(pkt)
	if err != nil {
		t.Fatal(err)
	}
	if h.Src != src || h.Dst != dst || h.Protocol != ip.ProtoTCP {
		t.Fatalf("%+v", h)
	}
	gp := gopacket.NewPacket(pkt, layers.LayerTypeIPv4, gopacket.Default)
	ip4, ok := gp.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if !ok {
		t.Fatal(gp.ErrorLayer())
	}
	if ip4.Checksum != h.Checksum {
		t.Fatalf("checksum ours=%04x gopacket=%04x", h.Checksum, ip4.Checksum)
	}
}

func TestBadChecksum(t *testing.T) {
	pkt := ip.Marshal(netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("2.2.2.2"), 6, 64, 1, []byte{9})
	pkt[10] ^= 0xff
	if _, err := ip.Parse(pkt); err != ip.ErrChecksum {
		t.Fatalf("err=%v", err)
	}
}

func TestRandomVsGopacket(t *testing.T) {
	src := netip.MustParseAddr("192.0.2.1")
	dst := netip.MustParseAddr("198.51.100.2")
	for id := uint16(0); id < 256; id++ {
		p := make([]byte, int(id%40)+1)
		for i := range p {
			p[i] = byte(i + int(id))
		}
		pkt := ip.Marshal(src, dst, 6, 64, id, p)
		want := ip.HeaderChecksum(pkt[:20])
		h, err := ip.Parse(pkt)
		if err != nil {
			t.Fatal(err)
		}
		if h.Checksum != want {
			t.Fatalf("id %d", id)
		}
		gp := gopacket.NewPacket(pkt, layers.LayerTypeIPv4, gopacket.Default)
		ip4 := gp.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
		if ip4.Checksum != h.Checksum {
			t.Fatalf("gopacket mismatch id=%d", id)
		}
	}
}
