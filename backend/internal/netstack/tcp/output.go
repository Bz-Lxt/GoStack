package tcp

import (
	"time"

	"gostack/internal/netstack/seq"
)

func (c *Conn) sendCtl(flags uint16, seqn, ack uint32, payload []byte, reason string) {
	h := Header{
		SrcPort: c.quad.LocalPort,
		DstPort: c.quad.RemotePort,
		Seq:     seqn,
		Ack:     ack,
		Flags:   flags,
		Window:  c.recv.Window(),
	}
	if flags&FlagSYN != 0 {
		h.Options = EncodeMSS(uint16(c.mss))
	}
	_ = c.host.SendTCP(c.quad.LocalIP, c.quad.RemoteIP, h, payload)
	c.txBytes += uint64(len(payload))
	c.lastSend = c.host.Now()
	_ = reason
}

func (c *Conn) sendACKLocked(trigger string) {
	c.ackPending = false
	c.timers.Stop("dack")
	c.sendCtl(FlagACK, c.sndNxt, c.rcvNxt, nil, trigger)
}

func (c *Conn) scheduleAck() {
	c.ackPending = true
	c.timers.AfterFunc("dack", 200*time.Millisecond, func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.ackPending && !c.closed {
			c.sendACKLocked("delayed ack")
		}
	})
}

func (c *Conn) sendFINLocked(trigger string) {
	if c.finSent {
		return
	}
	c.finSeq = c.sndNxt
	c.finSent = true
	c.sndNxt++
	c.sendCtl(FlagACK|FlagFIN, c.finSeq, c.rcvNxt, nil, trigger)
	c.armRTO()
}

func (c *Conn) trySendLocked() {
	if !c.state.CanSendData() && c.state != SynRcvd && c.state != SynSent {
		return
	}
	usable := min32(c.cc.Usable(), c.sndWnd)
	inflight := uint32(seq.Sub(c.sndNxt, c.sndUna))
	if c.sndWnd == 0 {
		c.armPersist()
		return
	}
	c.timers.Stop("persist")
	if inflight >= usable {
		return
	}
	budget := usable - inflight
	unsent := c.send.Unsent()
	for budget > 0 && len(unsent) > 0 {
		n := int(min32(budget, c.mss))
		if n > len(unsent) {
			n = len(unsent)
		}
		chunk := append([]byte(nil), unsent[:n]...)
		seqn := c.sndNxt
		c.sendCtl(FlagACK|FlagPSH, seqn, c.rcvNxt, chunk, "data")
		if !c.rttPending {
			c.rttPending = true
			c.rttStart = c.host.Now()
			c.lastRTTSeq = seqn + uint32(n)
		}
		c.send.MarkSent(n)
		c.sndNxt += uint32(n)
		budget -= uint32(n)
		unsent = c.send.Unsent()
		c.armRTO()
	}
	c.emitWindow()
}

func (c *Conn) onRTO() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || !seq.Less(c.sndUna, c.sndNxt) {
		return
	}
	if c.rex.Last.Count >= c.host.MaxRetries() && c.rex.Last.Seq == c.sndUna {
		c.hardErr = ErrTimeout
		c.setStateLocked(Closed, "max retries")
		c.host.Emit(Event{Type: "rto.timeout", Precise: true, ConnID: c.ID(), Payload: map[string]any{
			"rto_ms": c.rto.Current.Milliseconds(), "srtt_ms": c.rto.SRTT.Milliseconds(), "rttvar_ms": c.rto.RTTVAR.Milliseconds(),
		}})
		return
	}
	c.rttPending = false
	c.cc.OnTimeout(uint32(seq.Sub(c.sndNxt, c.sndUna)))
	c.host.Emit(Event{Type: "congestion.change", Precise: true, ConnID: c.ID(), Payload: map[string]any{
		"phase": "timeout", "cwnd": c.cc.CWND, "ssthresh": c.cc.SSThresh,
	}})
	c.rto.Backoff()
	c.host.Emit(Event{Type: "rto.timeout", Precise: true, ConnID: c.ID(), Payload: map[string]any{
		"rto_ms": c.rto.Current.Milliseconds(), "srtt_ms": c.rto.SRTT.Milliseconds(), "rttvar_ms": c.rto.RTTVAR.Milliseconds(),
	}})
	c.retransmitLocked("rto")
	c.armRTO()
}

func (c *Conn) retransmitLocked(reason string) {
	n := int(c.mss)
	if c.finSent && c.sndUna == c.finSeq && !c.finAcked {
		c.sendCtl(FlagACK|FlagFIN, c.finSeq, c.rcvNxt, nil, "retransmit fin")
		e := c.rex.Record(c.finSeq, 1, c.host.Now(), reason)
		c.host.Emit(Event{Type: "retransmit", Precise: true, ConnID: c.ID(), Payload: map[string]any{
			"seq": e.Seq, "len": 1, "reason": reason, "count": e.Count,
		}})
		return
	}
	if c.state == SynSent || (c.state == SynRcvd && c.sndUna == c.iss) {
		flags := uint16(FlagSYN)
		if c.state == SynRcvd {
			flags |= FlagACK
		}
		c.sendCtl(flags, c.iss, c.rcvNxt, nil, "retransmit syn")
		return
	}
	payload := c.send.PeekFromUNA(n)
	if len(payload) == 0 {
		return
	}
	c.sendCtl(FlagACK|FlagPSH, c.sndUna, c.rcvNxt, append([]byte(nil), payload...), "retransmit")
	e := c.rex.Record(c.sndUna, len(payload), c.host.Now(), reason)
	c.host.Emit(Event{Type: "retransmit", Precise: true, ConnID: c.ID(), Payload: map[string]any{
		"seq": e.Seq, "len": e.Len, "reason": reason, "count": e.Count,
	}})
	c.emitWindow()
}

func (c *Conn) armRTO() {
	if !seq.Less(c.sndUna, c.sndNxt) {
		c.timers.Stop("rto")
		return
	}
	c.timers.AfterFunc("rto", c.rto.Current, c.onRTO)
}

func (c *Conn) armPersist() {
	c.timers.AfterFunc("persist", c.rto.Current, func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.closed || c.sndWnd != 0 {
			return
		}
		c.sendCtl(FlagACK, c.sndNxt, c.rcvNxt, nil, "zero window probe")
		c.armPersist()
	})
}

func min32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
