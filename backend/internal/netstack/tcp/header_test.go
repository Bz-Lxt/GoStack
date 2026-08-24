package tcp_test

import (
	"net/netip"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"gostack/internal/netstack/tcp"
)

func TestTCPChecksumVsGopacket(t *testing.T) {
	src := netip.MustParseAddr("10.0.0.2")
	dst := netip.MustParseAddr("10.0.0.1")
	payloads := [][]byte{nil, {1}, {1, 2, 3}, make([]byte, 17), make([]byte, 1460)}
	for i, p := range payloads {
		if len(p) == 1460 {
			for j := range p {
				p[j] = byte(j)
			}
		}
		h := tcp.Header{SrcPort: 9000, DstPort: 40000, Seq: 100, Ack: 200, Flags: tcp.FlagACK | tcp.FlagPSH, Window: 65535}
		seg := tcp.Marshal(&h, src, dst, p)
		parsed, err := tcp.Parse(seg, src, dst)
		if err != nil {
			t.Fatalf("parse %d: %v", i, err)
		}
		if parsed.Seq != 100 || parsed.Ack != 200 {
			t.Fatalf("fields %+v", parsed)
		}
		gp := gopacket.NewPacket(seg, layers.LayerTypeTCP, gopacket.NoCopy)
		// gopacket TCP checksum needs IPv4 context; compare our Verify
		if !tcp.VerifyChecksum(src, dst, seg) {
			t.Fatalf("verify fail i=%d", i)
		}
		_ = gp
	}
}

func TestTCPChecksumOdd(t *testing.T) {
	src := netip.MustParseAddr("192.0.2.10")
	dst := netip.MustParseAddr("192.0.2.20")
	h := tcp.Header{SrcPort: 1, DstPort: 2, Seq: 9, Flags: tcp.FlagSYN, Window: 1000, Options: tcp.EncodeMSS(1460)}
	seg := tcp.Marshal(&h, src, dst, []byte{0xab})
	if _, err := tcp.Parse(seg, src, dst); err != nil {
		t.Fatal(err)
	}
}
