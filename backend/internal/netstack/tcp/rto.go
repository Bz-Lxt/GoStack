package tcp

import "time"

// RFC 6298.

const (
	alphaN = 8
	betaN  = 4
	kRTO   = 4
)

type RTO struct {
	SRTT    time.Duration
	RTTVAR  time.Duration
	Current time.Duration
	Min     time.Duration
	Max     time.Duration
	G       time.Duration
	first   bool
}

func NewRTO(min, max time.Duration) *RTO {
	if min <= 0 {
		min = 200 * time.Millisecond
	}
	if max <= 0 {
		max = 60 * time.Second
	}
	return &RTO{
		Current: 1 * time.Second,
		Min:     min,
		Max:     max,
		G:       10 * time.Millisecond,
	}
}

func (r *RTO) Sample(measured time.Duration) {
	if measured <= 0 {
		return
	}
	if !r.first {
		r.SRTT = measured
		r.RTTVAR = measured / 2
		r.first = true
	} else {
		diff := r.SRTT - measured
		if diff < 0 {
			diff = -diff
		}
		r.RTTVAR = r.RTTVAR*(betaN-1)/betaN + diff/betaN
		r.SRTT = r.SRTT*(alphaN-1)/alphaN + measured/alphaN
	}
	r.recompute()
}

func (r *RTO) recompute() {
	g := r.G
	if 4*r.RTTVAR > g {
		g = 4 * r.RTTVAR
	}
	v := r.SRTT + g
	if v < r.Min {
		v = r.Min
	}
	if v > r.Max {
		v = r.Max
	}
	r.Current = v
}

func (r *RTO) Backoff() {
	v := r.Current * 2
	if v > r.Max {
		v = r.Max
	}
	r.Current = v
}

func (r *RTO) Snapshot() (rto, srtt, rttvar time.Duration) {
	return r.Current, r.SRTT, r.RTTVAR
}
