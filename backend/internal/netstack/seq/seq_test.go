package seq_test

import (
	"testing"

	"gostack/internal/netstack/seq"
)

func TestWrapCompare(t *testing.T) {
	a := uint32(0xfffffff0)
	b := uint32(0x10)
	if !seq.Less(a, b) {
		t.Fatal("wrapped a should be less than b")
	}
	if seq.Less(b, a) {
		t.Fatal("b should not be less than wrapped a")
	}
	if seq.Sub(b, a) != 0x20 {
		t.Fatalf("sub=%d", seq.Sub(b, a))
	}
}

func TestInWindowWrap(t *testing.T) {
	start := uint32(0xfffffff8)
	if !seq.InWindow(0x4, start, 16) {
		t.Fatal("0x4 should be in window")
	}
	if seq.InWindow(0x20, start, 16) {
		t.Fatal("0x20 should be outside")
	}
}

func TestAddGreater(t *testing.T) {
	if seq.Add(0xfffffffe, 3) != 1 {
		t.Fatal("add wrap")
	}
	if !seq.Greater(1, 0xffffff00) || !seq.GreaterEq(1, 1) || !seq.LessEq(1, 1) {
		t.Fatal("cmp")
	}
}

func TestOverlap(t *testing.T) {
	if !seq.Overlap(10, 5, 12, 10) {
		t.Fatal("should overlap")
	}
	if seq.Overlap(0, 5, 12, 10) {
		t.Fatal("should not overlap")
	}
}
