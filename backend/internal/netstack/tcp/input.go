package tcp

import (
	"gostack/internal/netstack/seq"
)

func (c *Conn) Handle(th *Header, payload []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed && c.state == Closed {
		return
	}
	if c.state == SynSent {
		c.handleSynSent(th, payload)
		return
	}
	c.handleGeneral(th, payload)
}

func (c *Conn) handleSynSent(th *Header, payload []byte) {
	if th.Has(FlagACK) {
		if !seq.Greater(th.Ack, c.iss) || seq.Greater(th.Ack, c.sndNxt) {
			if !th.Has(FlagRST) {
				c.sendCtl(FlagRST, th.Ack, 0, nil, "bad ack in syn-sent")
			}
			return
		}
	}
	if th.Has(FlagRST) {
		if th.Has(FlagACK) {
			c.hardErr = ErrRefused
			c.setStateLocked(Closed, "rcv RST")
		}
		return
	}
	if !th.Has(FlagSYN) {
		return
	}
	c.irs = th.Seq
	c.rcvNxt = th.Seq + 1
	c.recv.Init(c.rcvNxt)
	if th.Has(FlagACK) {
		c.sndUna = th.Ack
		c.send.Ack(th.Ack)
	}
	c.sndWnd = uint32(th.Window)
	c.sndWL1 = th.Seq
	c.sndWL2 = th.Ack
	opt := ParseOptions(th.Options)
	if opt.HasMSS && opt.MSS > 0 {
		c.mss = uint32(opt.MSS)
		c.cc.MSS = c.mss
	}
	if seq.Greater(c.sndUna, c.iss) {
		c.sendACKLocked("handshake ack")
		c.setStateLocked(Established, "rcv SYN,ACK")
		c.host.Emit(Event{Type: "conn.open", Precise: true, ConnID: c.ID(), Payload: map[string]any{
			"local": c.quad.Local(), "remote": c.quad.Remote(), "role": "active",
		}})
		c.trySendLocked()
		if len(payload) > 0 {
			c.acceptData(th, payload)
		}
		return
	}
	c.sendCtl(FlagSYN|FlagACK, c.iss, c.rcvNxt, nil, "simultaneous open")
	c.setStateLocked(SynRcvd, "rcv SYN")
}

func (c *Conn) handleGeneral(th *Header, payload []byte) {
	seglen := uint32(len(payload))
	if th.Has(FlagSYN) {
		seglen++
	}
	if th.Has(FlagFIN) {
		seglen++
	}
	if !c.acceptable(th.Seq, seglen) && c.state != Listen {
		if !th.Has(FlagRST) {
			c.sendACKLocked("unacceptable seq")
		}
		return
	}
	if th.Has(FlagRST) {
		c.hardErr = ErrReset
		c.setStateLocked(Closed, "rcv RST")
		return
	}
	if th.Has(FlagSYN) && c.state != SynRcvd {
		c.hardErr = ErrReset
		c.sendCtl(FlagRST|FlagACK, c.sndNxt, c.rcvNxt, nil, "unexpected SYN")
		c.setStateLocked(Closed, "unexpected SYN")
		return
	}
	if !th.Has(FlagACK) {
		return
	}
	if !c.processACK(th, payload) {
		return
	}
	c.updateWindow(th)
	if len(payload) > 0 && c.state.CanRecvData() || (c.state == SynRcvd && len(payload) > 0) {
		c.acceptData(th, payload)
	} else if len(payload) > 0 && c.state == Established {
		c.acceptData(th, payload)
	}
	if th.Has(FlagFIN) {
		c.processFIN(th, uint32(len(payload)))
	}
}

func (c *Conn) acceptable(segSeq, seglen uint32) bool {
	wnd := uint32(c.recv.Window())
	if wnd == 0 {
		return seglen == 0 && segSeq == c.rcvNxt
	}
	if seglen == 0 {
		return seq.InWindow(segSeq, c.rcvNxt, wnd)
	}
	return seq.Overlap(segSeq, seglen, c.rcvNxt, wnd)
}

func (c *Conn) processACK(th *Header, payload []byte) bool {
	if c.state == SynRcvd {
		if seq.Greater(th.Ack, c.iss) && seq.LessEq(th.Ack, c.sndNxt) {
			c.sndUna = th.Ack
			c.send.Ack(th.Ack)
			c.setStateLocked(Established, "rcv ACK")
			c.host.Emit(Event{Type: "conn.open", Precise: true, ConnID: c.ID(), Payload: map[string]any{
				"local": c.quad.Local(), "remote": c.quad.Remote(), "role": "passive",
			}})
			c.signal(c.sendSpace)
		} else {
			c.sendCtl(FlagRST, th.Ack, 0, nil, "bad ack syn-rcvd")
			return false
		}
	}
	if seq.LessEq(th.Ack, c.sndUna) {
		if th.Ack == c.sndUna && len(payload) == 0 && !th.Has(FlagFIN) && !th.Has(FlagSYN) {
			c.dupACK++
			if c.cc.Phase == PhaseFastRecovery {
				c.cc.InflateDupACK()
			} else if c.dupACK == 3 {
				c.cc.OnFastRetransmit(uint32(seq.Sub(c.sndNxt, c.sndUna)))
				c.host.Emit(Event{Type: "congestion.change", Precise: true, ConnID: c.ID(), Payload: map[string]any{
					"phase": "fast_recovery", "cwnd": c.cc.CWND, "ssthresh": c.cc.SSThresh,
				}})
				c.retransmitLocked("fast")
			}
		}
		return true
	}
	if seq.Greater(th.Ack, c.sndNxt) {
		c.sendACKLocked("ack too new")
		return false
	}
	acked := uint32(seq.Sub(th.Ack, c.sndUna))
	c.send.Ack(th.Ack)
	c.sndUna = th.Ack
	c.dupACK = 0
	if c.rttPending && seq.GreaterEq(th.Ack, c.lastRTTSeq) {
		c.rto.Sample(c.host.Now().Sub(c.rttStart))
		c.rttPending = false
	}
	if c.cc.Phase == PhaseFastRecovery {
		c.cc.OnFastRecoveryACK()
		c.host.Emit(Event{Type: "congestion.change", Precise: true, ConnID: c.ID(), Payload: map[string]any{
			"phase": c.cc.Phase.String(), "cwnd": c.cc.CWND, "ssthresh": c.cc.SSThresh,
		}})
	} else {
		old := c.cc.Phase
		c.cc.OnACK(acked)
		if old != c.cc.Phase {
			c.host.Emit(Event{Type: "congestion.change", Precise: true, ConnID: c.ID(), Payload: map[string]any{
				"phase": c.cc.Phase.String(), "cwnd": c.cc.CWND, "ssthresh": c.cc.SSThresh,
			}})
		}
	}
	if c.finSent && !c.finAcked && seq.Greater(th.Ack, c.finSeq) {
		c.finAcked = true
		switch c.state {
		case FinWait1:
			c.setStateLocked(FinWait2, "rcv ACK of FIN")
		case Closing:
			c.setStateLocked(TimeWait, "rcv ACK of FIN")
		case LastAck:
			c.setStateLocked(Closed, "rcv ACK of FIN")
			c.host.Emit(Event{Type: "conn.close", Precise: true, ConnID: c.ID(), Payload: map[string]any{
				"local": c.quad.Local(), "remote": c.quad.Remote(), "role": "passive",
			}})
		}
	}
	c.armRTO()
	c.signal(c.sendSpace)
	c.trySendLocked()
	c.emitWindow()
	return true
}

func (c *Conn) updateWindow(th *Header) {
	if seq.Less(c.sndWL1, th.Seq) || (c.sndWL1 == th.Seq && seq.LessEq(c.sndWL2, th.Ack)) {
		c.sndWnd = uint32(th.Window)
		c.sndWL1 = th.Seq
		c.sndWL2 = th.Ack
	}
}

func (c *Conn) acceptData(th *Header, payload []byte) {
	if th.Seq == c.rcvNxt {
		c.rcvNxt = c.recv.WriteInOrder(payload)
		more, nxt := c.ooo.PopReady(c.rcvNxt)
		if len(more) > 0 {
			c.rcvNxt = c.recv.WriteInOrder(more)
		}
		c.rcvNxt = nxt
		c.rxBytes += uint64(len(payload) + len(more))
		c.signal(c.dataAvail)
		c.scheduleAck()
		if len(payload) >= int(c.mss) {
			c.sendACKLocked("full segment")
		}
		c.emitWindow()
		return
	}
	if seq.Greater(th.Seq, c.rcvNxt) {
		c.ooo.Insert(th.Seq, payload)
		c.sendACKLocked("out of order")
		return
	}
	c.sendACKLocked("old data")
}

func (c *Conn) processFIN(th *Header, dataLen uint32) {
	finSeq := th.Seq + dataLen
	if finSeq != c.rcvNxt {
		return
	}
	if c.peerFin {
		c.sendACKLocked("dup fin")
		return
	}
	c.peerFin = true
	c.readClosed = true
	c.rcvNxt++
	c.sendACKLocked("rcv FIN")
	c.signal(c.dataAvail)
	switch c.state {
	case Established, SynRcvd:
		c.setStateLocked(CloseWait, "rcv FIN")
	case FinWait1:
		if c.finAcked {
			c.setStateLocked(TimeWait, "rcv FIN")
		} else {
			c.setStateLocked(Closing, "rcv FIN")
		}
	case FinWait2:
		c.setStateLocked(TimeWait, "rcv FIN")
	}
}

func (c *Conn) StartPassive(iss, irs uint32, wnd uint16, mss uint32) {
	c.role = "passive"
	c.iss = iss
	c.irs = irs
	c.sndUna = iss
	c.sndNxt = iss + 1
	c.rcvNxt = irs + 1
	c.sndWnd = uint32(wnd)
	c.send.Init(iss + 1)
	c.recv.Init(c.rcvNxt)
	if mss > 0 {
		c.mss = mss
		c.cc.MSS = mss
	}
	c.state = SynRcvd
	c.sendCtl(FlagSYN|FlagACK, c.iss, c.rcvNxt, nil, "snd SYN,ACK")
	c.armRTO()
	c.host.Emit(Event{
		Type: "state.transition", Precise: true, ConnID: c.ID(),
		Payload: map[string]any{"from": "LISTEN", "to": "SYN_RCVD", "trigger": "rcv SYN / snd SYN,ACK"},
	})
}

func (c *Conn) StartActive(iss uint32) {
	c.role = "active"
	c.iss = iss
	c.sndUna = iss
	c.sndNxt = iss + 1
	c.send.Init(iss + 1)
	c.state = SynSent
	c.sendCtl(FlagSYN, c.iss, 0, nil, "snd SYN")
	c.armRTO()
	c.host.Emit(Event{
		Type: "state.transition", Precise: true, ConnID: c.ID(),
		Payload: map[string]any{"from": "CLOSED", "to": "SYN_SENT", "trigger": "active open / snd SYN"},
	})
}
