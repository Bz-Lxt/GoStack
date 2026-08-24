package device

import "sync"

var packetPool = sync.Pool{
	New: func() any {
		b := make([]byte, 2048)
		return &b
	},
}

func GetBuf() []byte {
	p := packetPool.Get().(*[]byte)
	return (*p)[:cap(*p)]
}

func PutBuf(b []byte) {
	if cap(b) < 2048 || cap(b) > 8192 {
		return
	}
	b = b[:cap(b)]
	packetPool.Put(&b)
}
