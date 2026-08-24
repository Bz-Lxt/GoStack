package tcp

import (
	"testing"
	"time"
)

func TestRTO6298(t *testing.T) {
	r := NewRTO(200*time.Millisecond, 60*time.Second)
	r.G = time.Millisecond
	r.Sample(200 * time.Millisecond)
	if r.SRTT != 200*time.Millisecond {
		t.Fatalf("srtt %v", r.SRTT)
	}
	if r.RTTVAR != 100*time.Millisecond {
		t.Fatalf("rttvar %v", r.RTTVAR)
	}
	// RTO = SRTT + max(G, 4*RTTVAR) = 200 + 400 = 600ms
	if r.Current != 600*time.Millisecond {
		t.Fatalf("rto %v", r.Current)
	}
	r.Sample(200 * time.Millisecond)
	// RTTVAR = 100*3/4 + 0/4 = 75
	// SRTT = 200*7/8 + 200/8 = 200
	// RTO = 200 + 300 = 500
	if r.SRTT != 200*time.Millisecond {
		t.Fatalf("srtt2 %v", r.SRTT)
	}
	if r.RTTVAR != 75*time.Millisecond {
		t.Fatalf("rttvar2 %v", r.RTTVAR)
	}
	if r.Current != 500*time.Millisecond {
		t.Fatalf("rto2 %v", r.Current)
	}
	before := r.Current
	r.Backoff()
	if r.Current != 2*before {
		t.Fatalf("backoff %v", r.Current)
	}
}

func TestRTOClamp(t *testing.T) {
	r := NewRTO(200*time.Millisecond, 60*time.Second)
	r.Sample(time.Millisecond)
	if r.Current < 200*time.Millisecond {
		t.Fatalf("min clamp %v", r.Current)
	}
}
