package tcp

import "gostack/internal/netstack/seq"

type SendBuffer struct {
	data   []byte
	unaSeq uint32
	nxtOff int
	cap    int
}

func NewSendBuffer(iss uint32, cap int) *SendBuffer {
	return &SendBuffer{unaSeq: iss, cap: cap}
}

func (b *SendBuffer) Init(iss uint32) {
	b.data = b.data[:0]
	b.unaSeq = iss
	b.nxtOff = 0
}

func (b *SendBuffer) Free() int {
	if b.cap <= len(b.data) {
		return 0
	}
	return b.cap - len(b.data)
}

func (b *SendBuffer) Len() int { return len(b.data) }

func (b *SendBuffer) InFlight() uint32 {
	return uint32(b.nxtOff)
}

func (b *SendBuffer) Unsent() []byte {
	if b.nxtOff >= len(b.data) {
		return nil
	}
	return b.data[b.nxtOff:]
}

func (b *SendBuffer) Write(p []byte) int {
	n := len(p)
	if n > b.Free() {
		n = b.Free()
	}
	if n <= 0 {
		return 0
	}
	b.data = append(b.data, p[:n]...)
	return n
}

func (b *SendBuffer) MarkSent(n int) {
	if n < 0 {
		return
	}
	b.nxtOff += n
	if b.nxtOff > len(b.data) {
		b.nxtOff = len(b.data)
	}
}

func (b *SendBuffer) PeekFromUNA(max int) []byte {
	if max <= 0 || len(b.data) == 0 {
		return nil
	}
	if max > len(b.data) {
		max = len(b.data)
	}
	return b.data[:max]
}

func (b *SendBuffer) Ack(ack uint32) int {
	if !seq.Greater(ack, b.unaSeq) {
		return 0
	}
	adv := int(uint32(seq.Sub(ack, b.unaSeq)))
	if adv > len(b.data) {
		adv = len(b.data)
	}
	b.data = b.data[adv:]
	b.nxtOff -= adv
	if b.nxtOff < 0 {
		b.nxtOff = 0
	}
	b.unaSeq = ack
	return adv
}

func (b *SendBuffer) UNA() uint32 { return b.unaSeq }

func (b *SendBuffer) NXT() uint32 { return b.unaSeq + uint32(b.nxtOff) }
