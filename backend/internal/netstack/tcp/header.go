package tcp

import (
	"encoding/binary"
	"errors"
	"net/netip"

	"gostack/internal/netstack/checksum"
)

const MinHeader = 20

const (
	FlagFIN = 0x01
	FlagSYN = 0x02
	FlagRST = 0x04
	FlagPSH = 0x08
	FlagACK = 0x10
	FlagURG = 0x20
	FlagECE = 0x40
	FlagCWR = 0x80
	FlagNS  = 0x100
)

var (
	ErrTruncated = errors.New("tcp: truncated header")
	ErrOffset    = errors.New("tcp: data offset < 5")
	ErrChecksum  = errors.New("tcp: checksum mismatch")
)

type Header struct {
	SrcPort    uint16
	DstPort    uint16
	Seq        uint32
	Ack        uint32
	DataOff    uint8
	Flags      uint16
	Window     uint16
	Checksum   uint16
	Urgent     uint16
	Options    []byte
	PayloadOff int
}

func (h *Header) Has(f uint16) bool { return h.Flags&f != 0 }

func (h *Header) FlagNames() []string {
	names := make([]string, 0, 6)
	order := []struct {
		f uint16
		s string
	}{
		{FlagFIN, "FIN"}, {FlagSYN, "SYN"}, {FlagRST, "RST"},
		{FlagPSH, "PSH"}, {FlagACK, "ACK"}, {FlagURG, "URG"},
	}
	for _, x := range order {
		if h.Has(x.f) {
			names = append(names, x.s)
		}
	}
	return names
}

func Parse(b []byte, src, dst netip.Addr) (*Header, error) {
	if len(b) < MinHeader {
		return nil, ErrTruncated
	}
	off := int(b[12] >> 4)
	if off < 5 {
		return nil, ErrOffset
	}
	hdrLen := off * 4
	if len(b) < hdrLen {
		return nil, ErrTruncated
	}
	h := &Header{
		SrcPort:    binary.BigEndian.Uint16(b[0:2]),
		DstPort:    binary.BigEndian.Uint16(b[2:4]),
		Seq:        binary.BigEndian.Uint32(b[4:8]),
		Ack:        binary.BigEndian.Uint32(b[8:12]),
		DataOff:    uint8(off),
		Flags:      uint16(b[12]&0x01)<<8 | uint16(b[13]),
		Window:     binary.BigEndian.Uint16(b[14:16]),
		Checksum:   binary.BigEndian.Uint16(b[16:18]),
		Urgent:     binary.BigEndian.Uint16(b[18:20]),
		PayloadOff: hdrLen,
	}
	if hdrLen > MinHeader {
		h.Options = append([]byte(nil), b[MinHeader:hdrLen]...)
	}
	if !VerifyChecksum(src, dst, b) {
		return nil, ErrChecksum
	}
	return h, nil
}

func (h *Header) Payload(seg []byte) []byte {
	if h.PayloadOff > len(seg) {
		return nil
	}
	return seg[h.PayloadOff:]
}

func Marshal(h *Header, src, dst netip.Addr, payload []byte) []byte {
	opt := padOptions(h.Options)
	hdrLen := MinHeader + len(opt)
	b := make([]byte, hdrLen+len(payload))
	binary.BigEndian.PutUint16(b[0:2], h.SrcPort)
	binary.BigEndian.PutUint16(b[2:4], h.DstPort)
	binary.BigEndian.PutUint32(b[4:8], h.Seq)
	binary.BigEndian.PutUint32(b[8:12], h.Ack)
	flags := h.Flags
	b[12] = byte((hdrLen/4)<<4) | byte((flags>>8)&0x01)
	b[13] = byte(flags)
	binary.BigEndian.PutUint16(b[14:16], h.Window)
	binary.BigEndian.PutUint16(b[18:20], h.Urgent)
	copy(b[MinHeader:hdrLen], opt)
	copy(b[hdrLen:], payload)
	sum := TCPChecksum(src, dst, b)
	binary.BigEndian.PutUint16(b[16:18], sum)
	h.Checksum = sum
	h.DataOff = uint8(hdrLen / 4)
	return b
}

func padOptions(opt []byte) []byte {
	if len(opt) == 0 {
		return nil
	}
	n := (len(opt) + 3) &^ 3
	out := make([]byte, n)
	copy(out, opt)
	return out
}

func pseudo(src, dst netip.Addr, tcpLen uint16) []byte {
	p := make([]byte, 12)
	sb, _ := src.MarshalBinary()
	db, _ := dst.MarshalBinary()
	copy(p[0:4], sb)
	copy(p[4:8], db)
	p[9] = 6
	binary.BigEndian.PutUint16(p[10:12], tcpLen)
	return p
}

func TCPChecksum(src, dst netip.Addr, tcp []byte) uint16 {
	tmp := make([]byte, len(tcp))
	copy(tmp, tcp)
	if len(tmp) >= 18 {
		tmp[16], tmp[17] = 0, 0
	}
	return checksum.Sum(pseudo(src, dst, uint16(len(tcp))), tmp)
}

func VerifyChecksum(src, dst netip.Addr, tcp []byte) bool {
	return checksum.Sum(pseudo(src, dst, uint16(len(tcp))), tcp) == 0
}
