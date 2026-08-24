package netstack

import (
	"context"
	"errors"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"gostack/internal/config"
	"gostack/internal/control"
	"gostack/internal/device"
	"gostack/internal/logger"
	"gostack/internal/netstack/ip"
	"gostack/internal/netstack/tcp"
	"gostack/internal/telemetry"
	"gostack/internal/timeutil"
)

type Stack struct {
	cfg   config.Config
	ip    netip.Addr
	dev   device.Device
	tab   *tcp.Table
	ports *tcp.PortAlloc
	lns   map[uint16]*tcp.Listener
	lnmu  sync.Mutex
	bus   *telemetry.Bus
	accum *telemetry.Accum
	stats *Stats
	impair *control.Impair
	step  *control.Step
	ipid  atomic.Uint32
	ctx   context.Context
	cancel context.CancelFunc
	wg    sync.WaitGroup
}

func New(cfg config.Config, dev device.Device, bus *telemetry.Bus, impair *control.Impair, step *control.Step) (*Stack, error) {
	addr, err := netip.ParseAddr(cfg.StackIP)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Stack{
		cfg: cfg, ip: addr, dev: dev,
		tab: tcp.NewTable(), ports: tcp.NewPortAlloc(),
		lns: map[uint16]*tcp.Listener{},
		bus: bus, accum: telemetry.NewAccum(), stats: &Stats{},
		impair: impair, step: step,
		ctx: ctx, cancel: cancel,
	}
	return s, nil
}

func (s *Stack) Start() {
	s.wg.Add(2)
	go s.rxLoop()
	go func() {
		defer s.wg.Done()
		telemetry.Tick(s.bus, s.accum, s.ctx.Done())
	}()
}

func (s *Stack) Close() error {
	s.cancel()
	_ = s.dev.Close()
	s.wg.Wait()
	return nil
}

func (s *Stack) DeviceName() string { return s.dev.Name() }
func (s *Stack) Bus() *telemetry.Bus { return s.bus }
func (s *Stack) Stats() *Stats       { return s.stats }
func (s *Stack) Accum() *telemetry.Accum { return s.accum }
func (s *Stack) IP() netip.Addr      { return s.ip }
func (s *Stack) Impair() *control.Impair { return s.impair }
func (s *Stack) Step() *control.Step { return s.step }

func (s *Stack) Now() time.Time          { return timeutil.Now() }
func (s *Stack) MSL() time.Duration      { return s.cfg.MSL }
func (s *Stack) MaxRetries() int         { return s.cfg.MaxRetries }
func (s *Stack) RTOMin() time.Duration   { return s.cfg.DemoRTOMin }

func (s *Stack) Emit(ev tcp.Event) {
	if ev.Type == "retransmit" {
		s.accum.AddRetrans()
	}
	s.bus.Publish(telemetry.New(ev.Type, ev.ConnID, ev.Precise, ev.Payload))
}

func (s *Stack) Remove(q tcp.Quad) {
	s.tab.Delete(q)
	s.stats.Conns.Store(int64(s.tab.Len()))
}

func (s *Stack) Connections() []tcp.Snapshot {
	list := s.tab.List()
	out := make([]tcp.Snapshot, 0, len(list))
	for _, c := range list {
		out = append(out, c.Snapshot())
	}
	return out
}

func (s *Stack) Connection(id string) (tcp.Snapshot, bool) {
	for _, c := range s.tab.List() {
		if c.ID() == id {
			return c.Snapshot(), true
		}
	}
	return tcp.Snapshot{}, false
}

func (s *Stack) Listen(addr string) (*tcp.Listener, error) {
	ipaddr, port, err := tcp.ParseListen(addr)
	if err != nil {
		return nil, err
	}
	if !ipaddr.IsValid() || !ipaddr.Is4() {
		ipaddr = s.ip
	}
	if !s.ports.Reserve(port) {
		return nil, errors.New("port in use")
	}
	ln := tcp.NewListener(s, tcp.Addr{IP: ipaddr, Port: port}, 128)
	s.lnmu.Lock()
	s.lns[port] = ln
	s.lnmu.Unlock()
	s.Emit(tcp.Event{Type: "state.transition", Precise: true, ConnID: "", Payload: map[string]any{
		"from": "CLOSED", "to": "LISTEN", "trigger": "passive open",
	}})
	return ln, nil
}

func (s *Stack) Dial(ctx context.Context, dst string) (*tcp.Conn, error) {
	rip, rport, err := tcp.ParseListen(dst)
	if err != nil {
		return nil, err
	}
	lp, ok := s.ports.Ephemeral()
	if !ok {
		return nil, errors.New("no ephemeral port")
	}
	q := tcp.Quad{LocalIP: s.ip, LocalPort: lp, RemoteIP: rip, RemotePort: rport}
	c := tcp.NewConn(s, q, 1460, s.cfg.SendBuf, s.cfg.RecvBuf)
	s.tab.Put(q, c)
	s.stats.Conns.Store(int64(s.tab.Len()))
	c.StartActive(tcp.GenerateISN(q))
	if ctx == nil {
		ctx = context.Background()
	}
	done := make(chan error, 1)
	go func() { done <- c.WaitEstablished(15 * time.Second) }()
	select {
	case <-ctx.Done():
		_ = c.Close()
		return nil, ctx.Err()
	case err := <-done:
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (s *Stack) listener(port uint16) *tcp.Listener {
	s.lnmu.Lock()
	defer s.lnmu.Unlock()
	return s.lns[port]
}

func (s *Stack) rxLoop() {
	defer s.wg.Done()
	buf := make([]byte, 2048)
	for {
		n, err := s.dev.Read(buf)
		if err != nil {
			if s.ctx.Err() != nil {
				return
			}
			logger.Guard().Warn("tun read", "err", err)
			time.Sleep(20 * time.Millisecond)
			continue
		}
		pkt := append([]byte(nil), buf[:n]...)
		s.handlePacket(pkt)
	}
}

func (s *Stack) handlePacket(pkt []byte) {
	defer func() {
		if rec := recover(); rec != nil {
			logger.Guard().Error("packet panic", "recover", rec)
		}
	}()
	s.stats.RxPackets.Add(1)
	s.stats.RxBytes.Add(uint64(len(pkt)))
	iph, err := ip.Parse(pkt)
	if err != nil {
		s.stats.IPBad.Add(1)
		if errors.Is(err, ip.ErrChecksum) {
			s.stats.ChecksumIP.Add(1)
			s.bus.Publish(telemetry.New("checksum.error", "", true, map[string]any{"layer": "ip"}))
		}
		return
	}
	if iph.Dst != s.ip {
		return
	}
	payload := iph.Payload(pkt)
	switch iph.Protocol {
	case ip.ProtoICMP:
		s.handleICMP(iph, payload)
	case ip.ProtoTCP:
		s.handleTCP(iph, payload)
	default:
		s.stats.DroppedProto.Add(1)
	}
}

func (s *Stack) handleICMP(iph *ip.Header, payload []byte) {
	if len(payload) < 8 || payload[0] != ip.ICMPEchoRequest {
		return
	}
	if _, ok := ip.ParseICMPEcho(payload); !ok {
		return
	}
	reply := ip.EchoReply(payload)
	out := ip.Marshal(s.ip, iph.Src, ip.ProtoICMP, 64, s.nextID(), reply)
	s.stats.ICMPEcho.Add(1)
	_, _ = s.dev.Write(out)
}

func (s *Stack) handleTCP(iph *ip.Header, payload []byte) {
	th, err := tcp.Parse(payload, iph.Src, iph.Dst)
	if err != nil {
		s.stats.TCPBad.Add(1)
		if errors.Is(err, tcp.ErrChecksum) {
			s.stats.ChecksumTCP.Add(1)
			s.bus.Publish(telemetry.New("checksum.error", "", true, map[string]any{"layer": "tcp"}))
		}
		return
	}
	data := th.Payload(payload)
	s.emitPacket("packet.rx", th.Has(tcp.FlagSYN)||th.Has(tcp.FlagFIN)||th.Has(tcp.FlagRST)||len(data) == 0, iph.Src, iph.Dst, th, data, "tun")
	s.accum.AddData(uint64(len(data)))
	q := tcp.Quad{LocalIP: iph.Dst, LocalPort: th.DstPort, RemoteIP: iph.Src, RemotePort: th.SrcPort}
	if c := s.tab.Get(q); c != nil {
		c.Handle(th, data)
		if c.TakePassiveReady() {
			if ln := s.listener(th.DstPort); ln != nil {
				if !ln.Enqueue(c) {
					c.Close()
				}
			}
		}
		return
	}
	if th.Has(tcp.FlagSYN) && !th.Has(tcp.FlagACK) {
		ln := s.listener(th.DstPort)
		if ln == nil || ln.Closed() {
			s.sendRST(iph.Dst, iph.Src, th, uint32(len(data))+1)
			return
		}
		opt := tcp.ParseOptions(th.Options)
		mss := uint32(1460)
		if opt.HasMSS && opt.MSS > 0 {
			mss = uint32(opt.MSS)
		}
		c := tcp.NewConn(s, q, mss, s.cfg.SendBuf, s.cfg.RecvBuf)
		s.tab.Put(q, c)
		s.stats.Conns.Store(int64(s.tab.Len()))
		c.StartPassive(tcp.GenerateISN(q), th.Seq, th.Window, mss)
		return
	}
	seglen := uint32(len(data))
	if th.Has(tcp.FlagFIN) || th.Has(tcp.FlagSYN) {
		seglen++
	}
	s.sendRST(iph.Dst, iph.Src, th, seglen)
}

func (s *Stack) emitPacket(typ string, precise bool, src, dst netip.Addr, th *tcp.Header, payload []byte, layer string) {
	hdr := tcp.Marshal(&tcp.Header{
		SrcPort: th.SrcPort, DstPort: th.DstPort, Seq: th.Seq, Ack: th.Ack,
		Flags: th.Flags, Window: th.Window, Options: th.Options,
	}, src, dst, payload)
	id := tcp.Quad{LocalIP: dst, LocalPort: th.DstPort, RemoteIP: src, RemotePort: th.SrcPort}.String()
	if typ == "packet.tx" {
		id = tcp.Quad{LocalIP: src, LocalPort: th.SrcPort, RemoteIP: dst, RemotePort: th.DstPort}.String()
	}
	s.bus.Publish(telemetry.New(typ, id, precise, map[string]any{
		"dir": map[bool]string{true: "tx", false: "rx"}[typ == "packet.tx"],
		"src": src.String() + ":" + itoa(th.SrcPort),
		"dst": dst.String() + ":" + itoa(th.DstPort),
		"seq": th.Seq, "ack": th.Ack, "flags": th.FlagNames(),
		"wnd": th.Window, "len": len(payload), "hex": encodeHex(hdr, 64),
		"layer": layer,
	}))
}

func itoa(v uint16) string {
	if v == 0 {
		return "0"
	}
	var b [6]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}
