package tcp

type RecvBuffer struct {
	data   []byte
	rcvNxt uint32
	cap    int
}

func NewRecvBuffer(nxt uint32, cap int) *RecvBuffer {
	return &RecvBuffer{rcvNxt: nxt, cap: cap}
}

func (b *RecvBuffer) Init(nxt uint32) {
	b.data = b.data[:0]
	b.rcvNxt = nxt
}

func (b *RecvBuffer) Available() int { return len(b.data) }

func (b *RecvBuffer) Window() uint16 {
	used := len(b.data)
	if used >= b.cap {
		return 0
	}
	w := b.cap - used
	if w > 65535 {
		return 65535
	}
	return uint16(w)
}

func (b *RecvBuffer) WriteInOrder(p []byte) uint32 {
	if len(p) == 0 {
		return b.rcvNxt
	}
	room := b.cap - len(b.data)
	if room <= 0 {
		return b.rcvNxt
	}
	if len(p) > room {
		p = p[:room]
	}
	b.data = append(b.data, p...)
	b.rcvNxt += uint32(len(p))
	return b.rcvNxt
}

func (b *RecvBuffer) Read(p []byte) int {
	n := copy(p, b.data)
	if n > 0 {
		b.data = b.data[n:]
	}
	return n
}

func (b *RecvBuffer) NXT() uint32 { return b.rcvNxt }
