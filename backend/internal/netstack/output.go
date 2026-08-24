package netstack

import (
	"encoding/binary"
	"net/netip"
	"sync/atomic"

	"gostack/internal/netstack/ip"
	"gostack/internal/netstack/tcp"
)

func (s *Stack) nextID() uint16 {
	return uint16(s.ipid.Add(1))
}

func (s *Stack) SendTCP(src, dst netip.Addr, hdr tcp.Header, payload []byte) error {
	if s.step != nil {
		s.step.Wait()
	}
	seg := tcp.Marshal(&hdr, src, dst, payload)
	pkt := ip.Marshal(src, dst, ip.ProtoTCP, 64, s.nextID(), seg)
	if s.impair != nil && s.impair.Drop() {
		return nil
	}
	if s.impair != nil {
		pkt = s.impair.MaybeDelay(pkt)
		if pkt == nil {
			return nil
		}
	}
	s.stats.TxPackets.Add(1)
	s.stats.TxBytes.Add(uint64(len(pkt)))
	precise := hdr.Has(tcp.FlagSYN) || hdr.Has(tcp.FlagFIN) || hdr.Has(tcp.FlagRST) || len(payload) == 0
	s.emitPacket("packet.tx", precise, src, dst, &hdr, payload, "tun")
	_, err := s.dev.Write(pkt)
	return err
}

func (s *Stack) sendRST(src, dst netip.Addr, th *tcp.Header, seglen uint32) {
	h := tcp.Header{
		SrcPort: th.DstPort,
		DstPort: th.SrcPort,
		Flags:   tcp.FlagRST,
	}
	if th.Has(tcp.FlagACK) {
		h.Seq = th.Ack
	} else {
		h.Flags |= tcp.FlagACK
		h.Ack = th.Seq + seglen
	}
	s.stats.RSTOut.Add(1)
	_ = s.SendTCP(src, dst, h, nil)
}

func encodeHex(b []byte, n int) string {
	if len(b) > n {
		b = b[:n]
	}
	const hexd = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = hexd[v>>4]
		out[i*2+1] = hexd[v&0x0f]
	}
	return string(out)
}

func u16(b []byte) uint16 { return binary.BigEndian.Uint16(b) }

var _ = atomic.Uint32{}
