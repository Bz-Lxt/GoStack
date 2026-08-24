package ip

import (
	"encoding/binary"
	"errors"
	"net/netip"

	"gostack/internal/netstack/checksum"
)

const (
	Version4   = 4
	MinHeader  = 20
	ProtoICMP  = 1
	ProtoTCP   = 6
	FlagDF     = 0x4000
	FlagMF     = 0x2000
	FragOffset = 0x1fff
)

var (
	ErrTruncated = errors.New("ipv4: truncated header")
	ErrVersion   = errors.New("ipv4: not version 4")
	ErrIHL       = errors.New("ipv4: ihl < 5")
	ErrLength    = errors.New("ipv4: total length mismatch")
	ErrChecksum  = errors.New("ipv4: checksum mismatch")
)

type Header struct {
	Version    uint8
	IHL        uint8
	TOS        uint8
	TotalLen   uint16
	ID         uint16
	FlagsFrag  uint16
	TTL        uint8
	Protocol   uint8
	Checksum   uint16
	Src        netip.Addr
	Dst        netip.Addr
	Options    []byte
	PayloadOff int
}

func Parse(b []byte) (*Header, error) {
	if len(b) < MinHeader {
		return nil, ErrTruncated
	}
	ihl := b[0] & 0x0f
	ver := b[0] >> 4
	if ver != Version4 {
		return nil, ErrVersion
	}
	if ihl < 5 {
		return nil, ErrIHL
	}
	hdrLen := int(ihl) * 4
	if len(b) < hdrLen {
		return nil, ErrTruncated
	}
	total := binary.BigEndian.Uint16(b[2:4])
	if int(total) < hdrLen || int(total) > len(b) {
		return nil, ErrLength
	}
	h := &Header{
		Version:    ver,
		IHL:        ihl,
		TOS:        b[1],
		TotalLen:   total,
		ID:         binary.BigEndian.Uint16(b[4:6]),
		FlagsFrag:  binary.BigEndian.Uint16(b[6:8]),
		TTL:        b[8],
		Protocol:   b[9],
		Checksum:   binary.BigEndian.Uint16(b[10:12]),
		Src:        mustIPv4(b[12:16]),
		Dst:        mustIPv4(b[16:20]),
		PayloadOff: hdrLen,
	}
	if hdrLen > MinHeader {
		h.Options = append([]byte(nil), b[MinHeader:hdrLen]...)
	}
	if !checksum.Verify(b[:hdrLen]) {
		return nil, ErrChecksum
	}
	return h, nil
}

func (h *Header) Payload(pkt []byte) []byte {
	end := int(h.TotalLen)
	if end > len(pkt) {
		end = len(pkt)
	}
	if h.PayloadOff > end {
		return nil
	}
	return pkt[h.PayloadOff:end]
}

func (h *Header) MoreFragments() bool { return h.FlagsFrag&FlagMF != 0 }

func (h *Header) FragOff() uint16 { return h.FlagsFrag & FragOffset }

func Marshal(src, dst netip.Addr, proto uint8, ttl uint8, id uint16, payload []byte) []byte {
	total := MinHeader + len(payload)
	b := make([]byte, total)
	b[0] = (Version4 << 4) | 5
	binary.BigEndian.PutUint16(b[2:4], uint16(total))
	binary.BigEndian.PutUint16(b[4:6], id)
	binary.BigEndian.PutUint16(b[6:8], FlagDF)
	b[8] = ttl
	b[9] = proto
	sb, _ := src.MarshalBinary()
	db, _ := dst.MarshalBinary()
	copy(b[12:16], sb)
	copy(b[16:20], db)
	copy(b[MinHeader:], payload)
	sum := checksum.Sum(b[:MinHeader])
	binary.BigEndian.PutUint16(b[10:12], sum)
	return b
}

func HeaderChecksum(hdr []byte) uint16 {
	if len(hdr) < MinHeader {
		return 0
	}
	tmp := make([]byte, len(hdr))
	copy(tmp, hdr)
	tmp[10], tmp[11] = 0, 0
	return checksum.Sum(tmp)
}

func mustIPv4(b []byte) netip.Addr {
	var a [4]byte
	copy(a[:], b)
	return netip.AddrFrom4(a)
}
