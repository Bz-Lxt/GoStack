package tcp

import (
	"net/netip"
	"testing"
	"time"
)

type fakeHost struct {
	sent []Header
	now  time.Time
}

func (f *fakeHost) SendTCP(_ netip.Addr, _ netip.Addr, hdr Header, _ []byte) error {
	f.sent = append(f.sent, hdr)
	return nil
}
func (f *fakeHost) Remove(Quad)            {}
func (f *fakeHost) Now() time.Time         { return f.now }
func (f *fakeHost) MSL() time.Duration     { return 20 * time.Millisecond }
func (f *fakeHost) MaxRetries() int        { return 5 }
func (f *fakeHost) RTOMin() time.Duration  { return 20 * time.Millisecond }
func (f *fakeHost) Emit(Event)             {}

func quad() Quad {
	return Quad{
		LocalIP: netip.MustParseAddr("10.0.0.2"), LocalPort: 9000,
		RemoteIP: netip.MustParseAddr("10.0.0.1"), RemotePort: 40000,
	}
}

func TestPassiveHandshakeAndCloseWait(t *testing.T) {
	h := &fakeHost{now: time.Now()}
	c := NewConn(h, quad(), 1460, 4096, 4096)
	c.StartPassive(1000, 2000, 4000, 1460)
	if c.state != SynRcvd {
		t.Fatal(c.state)
	}
	c.Handle(&Header{Seq: 2001, Ack: 1001, Flags: FlagACK, Window: 4000}, nil)
	if c.state != Established {
		t.Fatal(c.state)
	}
	if !c.TakePassiveReady() {
		t.Fatal("deliver")
	}
	c.Handle(&Header{Seq: 2001, Ack: 1001, Flags: FlagACK | FlagPSH, Window: 4000}, []byte("hi"))
	buf := make([]byte, 2)
	n, _ := c.Read(buf)
	if string(buf[:n]) != "hi" {
		t.Fatal(string(buf))
	}
	c.Handle(&Header{Seq: 2003, Ack: 1001, Flags: FlagACK | FlagFIN, Window: 4000}, nil)
	if c.state != CloseWait {
		t.Fatal(c.state)
	}
	_ = c.Close()
	if c.state != LastAck {
		t.Fatal(c.state)
	}
	c.Handle(&Header{Seq: 2004, Ack: c.finSeq + 1, Flags: FlagACK, Window: 4000}, nil)
	if c.state != Closed {
		t.Fatal(c.state)
	}
}

func TestActiveHandshakeAndFinWait(t *testing.T) {
	h := &fakeHost{now: time.Now()}
	c := NewConn(h, quad(), 1460, 4096, 4096)
	c.StartActive(5000)
	c.Handle(&Header{Seq: 8000, Ack: 5001, Flags: FlagSYN | FlagACK, Window: 8000, Options: EncodeMSS(1460)}, nil)
	if c.state != Established {
		t.Fatal(c.state)
	}
	if _, err := c.Write([]byte("xy")); err != nil {
		t.Fatal(err)
	}
	_ = c.Close()
	if c.state != FinWait1 {
		t.Fatal(c.state)
	}
	c.Handle(&Header{Seq: 8001, Ack: c.finSeq + 1, Flags: FlagACK, Window: 8000}, nil)
	if c.state != FinWait2 {
		t.Fatal(c.state)
	}
	c.Handle(&Header{Seq: 8001, Ack: c.finSeq + 1, Flags: FlagACK | FlagFIN, Window: 8000}, nil)
	if c.state != TimeWait {
		t.Fatal(c.state)
	}
}

func TestClosingSimultaneous(t *testing.T) {
	h := &fakeHost{now: time.Now()}
	c := NewConn(h, quad(), 1460, 4096, 4096)
	c.StartPassive(1, 2, 1000, 1460)
	c.Handle(&Header{Seq: 3, Ack: 2, Flags: FlagACK, Window: 1000}, nil)
	_ = c.Close()
	if c.state != FinWait1 {
		t.Fatal(c.state)
	}
	c.Handle(&Header{Seq: 3, Ack: 2, Flags: FlagACK | FlagFIN, Window: 1000}, nil)
	if c.state != Closing {
		t.Fatal(c.state)
	}
	c.Handle(&Header{Seq: 4, Ack: c.finSeq + 1, Flags: FlagACK, Window: 1000}, nil)
	if c.state != TimeWait {
		t.Fatal(c.state)
	}
}

func TestListenAndSynSentNames(t *testing.T) {
	if Listen.String() != "LISTEN" || SynSent.String() != "SYN_SENT" {
		t.Fatal("names")
	}
}
