package tcp

import "time"

type RexmitEntry struct {
	Seq      uint32
	Len      int
	First    time.Time
	Last     time.Time
	Count    int
	TimedOut bool
}

type RexmitLog struct {
	Last RexmitEntry
	N    int
}

func (l *RexmitLog) Record(seq uint32, n int, now time.Time, reason string) RexmitEntry {
	e := RexmitEntry{Seq: seq, Len: n, First: now, Last: now, Count: 1}
	if l.Last.Seq == seq {
		e.Count = l.Last.Count + 1
		e.First = l.Last.First
	}
	if reason == "rto" {
		e.TimedOut = true
	}
	l.Last = e
	l.N++
	return e
}
