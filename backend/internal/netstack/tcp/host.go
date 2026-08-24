package tcp

import (
	"net/netip"
	"time"
)

type Event struct {
	Type    string
	Precise bool
	ConnID  string
	Payload map[string]any
}

type Host interface {
	SendTCP(src, dst netip.Addr, hdr Header, payload []byte) error
	Remove(q Quad)
	Now() time.Time
	MSL() time.Duration
	MaxRetries() int
	RTOMin() time.Duration
	Emit(Event)
}

type Snapshot struct {
	ID        string    `json:"id"`
	Local     string    `json:"local"`
	Remote    string    `json:"remote"`
	State     string    `json:"state"`
	AliveMS   int64     `json:"alive_ms"`
	RxBytes   uint64    `json:"rx_bytes"`
	TxBytes   uint64    `json:"tx_bytes"`
	CWND      uint32    `json:"cwnd"`
	RWND      uint32    `json:"rwnd"`
	RTOMS     int64     `json:"rto_ms"`
	SRTTMS    int64     `json:"srtt_ms"`
	UNA       uint32    `json:"una"`
	NXT       uint32    `json:"nxt"`
	RcvNXT    uint32    `json:"rcv_nxt"`
	SSThresh  uint32    `json:"ssthresh"`
	Phase     string    `json:"phase"`
	DupACK    int       `json:"dup_acks"`
	Retrans   int       `json:"retransmits"`
	Opened    time.Time `json:"opened"`
}
