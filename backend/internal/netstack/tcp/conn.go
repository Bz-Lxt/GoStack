package tcp

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"gostack/internal/netstack/seq"
)

var (
	ErrClosed    = errors.New("tcp: connection closed")
	ErrReset     = errors.New("tcp: connection reset")
	ErrTimeout   = errors.New("tcp: i/o timeout")
	ErrRefused   = errors.New("tcp: connection refused")
	errWouldWait = errors.New("tcp: wait")
)

type Conn struct {
	mu sync.Mutex

	host Host
	quad Quad
	mss  uint32

	state State
	iss   uint32
	irs   uint32

	sndUna uint32
	sndNxt uint32
	sndWnd uint32
	sndWL1 uint32
	sndWL2 uint32

	rcvNxt uint32
	rcvWnd uint32

	send *SendBuffer
	recv *RecvBuffer
	ooo  OutOfOrderQueue
	cc   *Congestion
	rto  *RTO
	rex  RexmitLog

	finSeq     uint32
	finSent    bool
	finAcked   bool
	peerFin    bool
	readClosed bool
	hardErr    error

	dupACK     int
	rxBytes    uint64
	txBytes    uint64
	opened     time.Time
	lastRTTSeq uint32
	lastSend   time.Time
	rttPending bool
	rttStart   time.Time
	role       string
	delivered  bool

	ackPending bool
	probePend  bool

	dataAvail  chan struct{}
	sendSpace  chan struct{}
	established chan struct{}
	dead        chan struct{}
	onceDead   sync.Once

	readDL  time.Time
	writeDL time.Time

	timers *TimerWheel
	closed bool
}

func NewConn(host Host, q Quad, mss uint32, sendCap, recvCap int) *Conn {
	if mss == 0 {
		mss = 1460
	}
	c := &Conn{
		host:        host,
		quad:        q,
		mss:         mss,
		state:       Closed,
		send:        NewSendBuffer(0, sendCap),
		recv:        NewRecvBuffer(0, recvCap),
		cc:          NewCongestion(mss),
		rto:         NewRTO(host.RTOMin(), 60*time.Second),
		dataAvail:   make(chan struct{}, 1),
		sendSpace:   make(chan struct{}, 1),
		established: make(chan struct{}),
		dead:        make(chan struct{}),
		opened:      host.Now(),
		timers:      NewTimerWheel(),
	}
	return c
}

func (c *Conn) ID() string { return c.quad.String() }

func (c *Conn) LocalAddr() net.Addr  { return Addr{IP: c.quad.LocalIP, Port: c.quad.LocalPort} }
func (c *Conn) RemoteAddr() net.Addr { return Addr{IP: c.quad.RemoteIP, Port: c.quad.RemotePort} }

func (c *Conn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readDL = t
	c.writeDL = t
	c.kick()
	return nil
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readDL = t
	c.kick()
	return nil
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeDL = t
	c.kick()
	return nil
}

func (c *Conn) Read(p []byte) (int, error) {
	for {
		c.mu.Lock()
		if n := c.recv.Read(p); n > 0 {
			c.mu.Unlock()
			c.signal(c.sendSpace)
			return n, nil
		}
		err := c.hardErr
		eof := c.readClosed
		dl := c.readDL
		c.mu.Unlock()
		if err != nil {
			return 0, err
		}
		if eof {
			return 0, io.EOF
		}
		if err := c.wait(c.dataAvail, dl); err != nil {
			return 0, err
		}
	}
}

func (c *Conn) Write(p []byte) (int, error) {
	off := 0
	for off < len(p) {
		c.mu.Lock()
		if c.hardErr != nil {
			err := c.hardErr
			c.mu.Unlock()
			return off, err
		}
		if !c.state.CanSendData() && c.state != SynSent && c.state != SynRcvd {
			c.mu.Unlock()
			return off, ErrClosed
		}
		n := c.send.Write(p[off:])
		if n > 0 {
			off += n
			c.trySendLocked()
			c.mu.Unlock()
			continue
		}
		dl := c.writeDL
		c.mu.Unlock()
		if err := c.wait(c.sendSpace, dl); err != nil {
			return off, err
		}
	}
	return off, nil
}

func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	switch c.state {
	case Closed, Listen, TimeWait:
		c.abort(nil)
		return nil
	case SynSent:
		c.abort(ErrClosed)
		return nil
	case CloseWait:
		c.sendFINLocked("app close")
		c.setStateLocked(LastAck, "app close")
	default:
		if c.state == Established || c.state == SynRcvd {
			c.sendFINLocked("app close")
			c.setStateLocked(FinWait1, "app close")
		}
	}
	return nil
}

func (c *Conn) wait(ch <-chan struct{}, dl time.Time) error {
	if !dl.IsZero() && !time.Now().Before(dl) {
		return ErrTimeout
	}
	var timeout <-chan time.Time
	if !dl.IsZero() {
		t := time.NewTimer(time.Until(dl))
		defer t.Stop()
		timeout = t.C
	}
	select {
	case <-ch:
		return nil
	case <-c.dead:
		c.mu.Lock()
		err := c.hardErr
		c.mu.Unlock()
		if err != nil {
			return err
		}
		return ErrClosed
	case <-timeout:
		return ErrTimeout
	}
}

func (c *Conn) signal(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

func (c *Conn) kick() {
	c.signal(c.dataAvail)
	c.signal(c.sendSpace)
}

func (c *Conn) setStateLocked(to State, trigger string) {
	from := c.state
	if from == to {
		return
	}
	c.state = to
	c.host.Emit(Event{
		Type: "state.transition", Precise: true, ConnID: c.ID(),
		Payload: map[string]any{"from": from.String(), "to": to.String(), "trigger": trigger},
	})
	if to == Established {
		select {
		case <-c.established:
		default:
			close(c.established)
		}
	}
	if to == TimeWait {
		c.timers.AfterFunc("timewait", 2*c.host.MSL(), func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			if c.state == TimeWait {
				c.setStateLocked(Closed, "2MSL expired")
				c.abort(nil)
			}
		})
	}
	if to == Closed {
		c.abort(c.hardErr)
	}
}

func (c *Conn) abort(err error) {
	if c.hardErr == nil && err != nil {
		c.hardErr = err
	}
	c.closed = true
	c.readClosed = true
	c.timers.StopAll()
	c.onceDead.Do(func() { close(c.dead) })
	c.kick()
	c.host.Remove(c.quad)
}

func (c *Conn) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	rto, srtt, _ := c.rto.Snapshot()
	return Snapshot{
		ID:       c.ID(),
		Local:    c.quad.Local(),
		Remote:   c.quad.Remote(),
		State:    c.state.String(),
		AliveMS:  c.host.Now().Sub(c.opened).Milliseconds(),
		RxBytes:  c.rxBytes,
		TxBytes:  c.txBytes,
		CWND:     c.cc.CWND,
		RWND:     uint32(c.recv.Window()),
		RTOMS:    rto.Milliseconds(),
		SRTTMS:   srtt.Milliseconds(),
		UNA:      c.sndUna,
		NXT:      c.sndNxt,
		RcvNXT:   c.rcvNxt,
		SSThresh: c.cc.SSThresh,
		Phase:    c.cc.Phase.String(),
		DupACK:   c.dupACK,
		Retrans:  c.rex.N,
		Opened:   c.opened,
	}
}

func (c *Conn) WindowCells() []map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.windowCellsLocked()
}

func (c *Conn) windowCellsLocked() []map[string]any {
	cells := make([]map[string]any, 0, 32)
	mss := c.mss
	if mss == 0 {
		mss = 1460
	}
	base := c.sndUna
	right := c.sndUna + c.sndWnd
	if c.sndWnd == 0 {
		right = c.sndUna
	}
	for i := 0; i < 32; i++ {
		s := base + uint32(i)*mss
		mark := "outside"
		switch {
		case seq.Less(s, c.sndUna):
			mark = "acked"
		case !seq.GreaterEq(s, c.sndNxt):
			mark = "inflight"
		case c.sndWnd > 0 && seq.Less(s, right):
			mark = "usable"
		}
		if c.rex.Last.Count > 0 && c.rex.Last.Seq == s {
			mark = "retrans"
		}
		cells = append(cells, map[string]any{"seq": s, "len": mss, "mark": mark})
	}
	return cells
}

func (c *Conn) emitWindow() {
	c.host.Emit(Event{
		Type: "window.update", Precise: false, ConnID: c.ID(),
		Payload: map[string]any{
			"snd_una": c.sndUna, "snd_nxt": c.sndNxt, "snd_wnd": c.sndWnd,
			"rcv_nxt": c.rcvNxt, "rcv_wnd": uint32(c.recv.Window()),
			"cells": c.windowCellsLocked(),
		},
	})
}

func (c *Conn) TakePassiveReady() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state == Established && !c.delivered && c.role == "passive" {
		c.delivered = true
		return true
	}
	return false
}

func (c *Conn) WaitEstablished(d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-c.established:
		return nil
	case <-c.dead:
		c.mu.Lock()
		err := c.hardErr
		c.mu.Unlock()
		if err != nil {
			return err
		}
		return ErrClosed
	case <-t.C:
		return ErrTimeout
	}
}
