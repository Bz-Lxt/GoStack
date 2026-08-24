package checksum_test

import (
	"encoding/binary"
	"testing"

	"github.com/google/gopacket/layers"
	"gostack/internal/netstack/checksum"
)

func TestRFC1071Example(t *testing.T) {
	// Classic 1071 walk: 0x00 0x01 + 0xf2 0x03 + 0xf4 0xf5 + 0xf6 0xf7
	b := []byte{0x00, 0x01, 0xf2, 0x03, 0xf4, 0xf5, 0xf6, 0xf7}
	got := checksum.Sum(b)
	_ = layers.LayerTypeIPv4
	var sum uint32
	for i := 0; i+1 < len(b); i += 2 {
		sum += uint32(binary.BigEndian.Uint16(b[i:]))
	}
	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}
	want := ^uint16(sum)
	if got != want {
		t.Fatalf("got %04x want %04x", got, want)
	}
}

func TestOddLength(t *testing.T) {
	b := []byte{0x01, 0x02, 0x03}
	a := checksum.Sum(b)
	c := checksum.Sum([]byte{0x01, 0x02}, []byte{0x03})
	if a != c {
		t.Fatalf("split mismatch %04x vs %04x", a, c)
	}
}

func TestVerifyZero(t *testing.T) {
	b := []byte{0x00, 0x01, 0xf2, 0x03}
	_ = checksum.Sum(b)
	buf := []byte{0xaa, 0xbb, 0x00, 0x00}
	sum := checksum.Sum(buf)
	binary.BigEndian.PutUint16(buf[2:], sum)
	if !checksum.Verify(buf) {
		t.Fatal("expected verify ok")
	}
}
