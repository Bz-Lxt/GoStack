package checksum

// RFC 1071 one's complement checksum.

type Adder struct {
	sum uint32
	odd byte
	has bool
}

func (a *Adder) Add(p []byte) {
	if len(p) == 0 {
		return
	}
	if a.has {
		a.sum += uint32(a.odd)<<8 | uint32(p[0])
		p = p[1:]
		a.has = false
	}
	for i := 0; i+1 < len(p); i += 2 {
		a.sum += uint32(p[i])<<8 | uint32(p[i+1])
	}
	if len(p)%2 == 1 {
		a.odd = p[len(p)-1]
		a.has = true
	}
}

func (a *Adder) Finish() uint16 {
	if a.has {
		a.sum += uint32(a.odd) << 8
		a.has = false
	}
	s := a.sum
	for s > 0xffff {
		s = (s >> 16) + (s & 0xffff)
	}
	return ^uint16(s)
}

func Sum(parts ...[]byte) uint16 {
	var a Adder
	for _, p := range parts {
		a.Add(p)
	}
	return a.Finish()
}

func Fold(sum uint32) uint16 {
	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}
	return uint16(sum)
}

func Verify(parts ...[]byte) bool {
	return Sum(parts...) == 0
}
