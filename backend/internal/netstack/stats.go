package netstack

import "sync/atomic"

type Stats struct {
	RxPackets      atomic.Uint64 `json:"-"`
	TxPackets      atomic.Uint64 `json:"-"`
	RxBytes        atomic.Uint64 `json:"-"`
	TxBytes        atomic.Uint64 `json:"-"`
	IPBad          atomic.Uint64 `json:"-"`
	TCPBad         atomic.Uint64 `json:"-"`
	ChecksumIP     atomic.Uint64 `json:"-"`
	ChecksumTCP    atomic.Uint64 `json:"-"`
	DroppedProto   atomic.Uint64 `json:"-"`
	RSTOut         atomic.Uint64 `json:"-"`
	ICMPEcho       atomic.Uint64 `json:"-"`
	Conns          atomic.Int64  `json:"-"`
	TelemetryDrop  atomic.Uint64 `json:"-"`
}

func (s *Stats) Snapshot() map[string]uint64 {
	return map[string]uint64{
		"rx_packets":     s.RxPackets.Load(),
		"tx_packets":     s.TxPackets.Load(),
		"rx_bytes":       s.RxBytes.Load(),
		"tx_bytes":       s.TxBytes.Load(),
		"ip_bad":         s.IPBad.Load(),
		"tcp_bad":        s.TCPBad.Load(),
		"checksum_ip":    s.ChecksumIP.Load(),
		"checksum_tcp":   s.ChecksumTCP.Load(),
		"dropped_proto":  s.DroppedProto.Load(),
		"rst_out":        s.RSTOut.Load(),
		"icmp_echo":      s.ICMPEcho.Load(),
		"conns":          uint64(s.Conns.Load()),
		"telemetry_drop": s.TelemetryDrop.Load(),
	}
}
