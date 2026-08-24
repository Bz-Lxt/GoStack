package ip

import (
	"encoding/binary"

	"gostack/internal/netstack/checksum"
)

const (
	ICMPEchoReply   = 0
	ICMPEchoRequest = 8
)

type ICMPEcho struct {
	Type     uint8
	Code     uint8
	Checksum uint16
	ID       uint16
	Seq      uint16
	Data     []byte
}

func ParseICMPEcho(b []byte) (*ICMPEcho, bool) {
	if len(b) < 8 {
		return nil, false
	}
	m := &ICMPEcho{
		Type:     b[0],
		Code:     b[1],
		Checksum: binary.BigEndian.Uint16(b[2:4]),
		ID:       binary.BigEndian.Uint16(b[4:6]),
		Seq:      binary.BigEndian.Uint16(b[6:8]),
		Data:     append([]byte(nil), b[8:]...),
	}
	if !checksum.Verify(b) {
		return nil, false
	}
	return m, true
}

func (m *ICMPEcho) Marshal() []byte {
	b := make([]byte, 8+len(m.Data))
	b[0] = m.Type
	b[1] = m.Code
	binary.BigEndian.PutUint16(b[4:6], m.ID)
	binary.BigEndian.PutUint16(b[6:8], m.Seq)
	copy(b[8:], m.Data)
	sum := checksum.Sum(b)
	binary.BigEndian.PutUint16(b[2:4], sum)
	return b
}

func EchoReply(req []byte) []byte {
	if len(req) < 8 {
		return nil
	}
	out := append([]byte(nil), req...)
	out[0] = ICMPEchoReply
	out[2], out[3] = 0, 0
	sum := checksum.Sum(out)
	binary.BigEndian.PutUint16(out[2:4], sum)
	return out
}
