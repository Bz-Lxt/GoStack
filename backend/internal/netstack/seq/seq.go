package seq

// Unsigned 32-bit sequence arithmetic (RFC 1982 / RFC 793).
// Comparison MUST use signed 32-bit difference, never raw < / >.

func Add(a uint32, n uint32) uint32 { return a + n }

func Sub(a, b uint32) int32 { return int32(a - b) }

func Less(a, b uint32) bool { return int32(a-b) < 0 }

func LessEq(a, b uint32) bool { return int32(a-b) <= 0 }

func Greater(a, b uint32) bool { return int32(a-b) > 0 }

func GreaterEq(a, b uint32) bool { return int32(a-b) >= 0 }

func InWindow(seq, start, size uint32) bool {
	if size == 0 {
		return false
	}
	off := uint32(int32(seq - start))
	return off < size
}

func Overlap(seq, slen, start, size uint32) bool {
	if slen == 0 {
		return size > 0 && !Less(seq, start) && Less(seq, start+size)
	}
	end := seq + slen
	wndEnd := start + size
	return Less(seq, wndEnd) && Less(start, end)
}
