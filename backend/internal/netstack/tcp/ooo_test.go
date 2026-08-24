package tcp

import "testing"

func TestOOOReorder(t *testing.T) {
	var q OutOfOrderQueue
	q.Insert(12, []byte("cd"))
	q.Insert(10, []byte("ab"))
	got, nxt := q.PopReady(10)
	if string(got) != "abcd" || nxt != 14 {
		t.Fatalf("got=%q nxt=%d", got, nxt)
	}
}

func TestOOOOverlap(t *testing.T) {
	var q OutOfOrderQueue
	q.Insert(10, []byte("abcd"))
	q.Insert(12, []byte("cdef"))
	got, nxt := q.PopReady(10)
	if string(got) != "abcdef" || nxt != 16 {
		t.Fatalf("got=%q nxt=%d segs=%d", got, nxt, q.Len())
	}
}
