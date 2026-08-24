package tcp

import "gostack/internal/netstack/seq"

type oooSeg struct {
	seq  uint32
	data []byte
}

type OutOfOrderQueue struct {
	segs []oooSeg
}

func (q *OutOfOrderQueue) Insert(seqn uint32, data []byte) {
	if len(data) == 0 {
		return
	}
	cp := append([]byte(nil), data...)
	n := oooSeg{seq: seqn, data: cp}
	out := make([]oooSeg, 0, len(q.segs)+1)
	inserted := false
	for _, s := range q.segs {
		if !inserted && seq.LessEq(n.seq, s.seq) {
			out = mergeAppend(out, n)
			inserted = true
		}
		out = mergeAppend(out, s)
	}
	if !inserted {
		out = mergeAppend(out, n)
	}
	q.segs = out
}

func mergeAppend(dst []oooSeg, s oooSeg) []oooSeg {
	if len(dst) == 0 {
		return append(dst, s)
	}
	last := &dst[len(dst)-1]
	lastEnd := last.seq + uint32(len(last.data))
	sEnd := s.seq + uint32(len(s.data))
	if seq.LessEq(s.seq, lastEnd) && seq.LessEq(last.seq, sEnd) {
		if seq.Less(s.seq, last.seq) {
			prefix := int(uint32(seq.Sub(last.seq, s.seq)))
			if prefix > 0 && prefix < len(s.data) {
				last.data = append(append([]byte(nil), s.data[:prefix]...), last.data...)
				last.seq = s.seq
			}
		}
		if seq.Greater(sEnd, lastEnd) {
			skip := int(uint32(seq.Sub(lastEnd, s.seq)))
			if skip < 0 {
				skip = 0
			}
			if skip < len(s.data) {
				last.data = append(last.data, s.data[skip:]...)
			}
		}
		return dst
	}
	return append(dst, s)
}

func (q *OutOfOrderQueue) PopReady(nxt uint32) ([]byte, uint32) {
	var out []byte
	for len(q.segs) > 0 {
		s := q.segs[0]
		end := s.seq + uint32(len(s.data))
		if seq.GreaterEq(s.seq, nxt) && s.seq != nxt {
			if !seq.LessEq(s.seq, nxt) {
				break
			}
		}
		if seq.GreaterEq(s.seq, nxt) && s.seq != nxt {
			break
		}
		if seq.Less(end, nxt) || end == nxt && len(s.data) == 0 {
			q.segs = q.segs[1:]
			continue
		}
		if seq.Less(s.seq, nxt) {
			skip := int(uint32(seq.Sub(nxt, s.seq)))
			if skip >= len(s.data) {
				q.segs = q.segs[1:]
				continue
			}
			out = append(out, s.data[skip:]...)
			nxt = end
			q.segs = q.segs[1:]
			continue
		}
		if s.seq != nxt {
			break
		}
		out = append(out, s.data...)
		nxt = end
		q.segs = q.segs[1:]
	}
	return out, nxt
}

func (q *OutOfOrderQueue) Len() int { return len(q.segs) }
