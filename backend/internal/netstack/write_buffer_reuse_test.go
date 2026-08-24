package netstack_test

import (
	"bytes"
	"context"
	"net/netip"
	"testing"
	"time"

	"gostack/internal/config"
	"gostack/internal/device"
	"gostack/internal/netstack"
	"gostack/internal/netstack/ip"
	"gostack/internal/netstack/tcp"
	"gostack/internal/telemetry"
)

type tcpPacket struct {
	ip      *ip.Header
	tcp     *tcp.Header
	payload []byte
}

func readTCPPacket(t *testing.T, d device.Device) tcpPacket {
	t.Helper()
	type result struct {
		packet []byte
		err    error
	}
	resultc := make(chan result, 1)
	go func() {
		buf := make([]byte, 2048)
		n, err := d.Read(buf)
		resultc <- result{packet: buf[:n], err: err}
	}()

	var res result
	select {
	case res = <-resultc:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for TCP packet")
	}
	if res.err != nil {
		t.Fatalf("read packet: %v", res.err)
	}
	iph, err := ip.Parse(res.packet)
	if err != nil {
		t.Fatalf("parse IP packet: %v", err)
	}
	segment := iph.Payload(res.packet)
	th, err := tcp.Parse(segment, iph.Src, iph.Dst)
	if err != nil {
		t.Fatalf("parse TCP segment: %v", err)
	}
	return tcpPacket{
		ip:      iph,
		tcp:     th,
		payload: append([]byte(nil), th.Payload(segment)...),
	}
}

func writeTCPPacket(t *testing.T, d device.Device, src, dst netip.Addr, h tcp.Header) {
	t.Helper()
	segment := tcp.Marshal(&h, src, dst, nil)
	packet := ip.Marshal(src, dst, ip.ProtoTCP, 64, 1, segment)
	n, err := d.Write(packet)
	if err != nil {
		t.Fatalf("write packet: %v", err)
	}
	if n != len(packet) {
		t.Fatalf("short packet write: got %d, want %d", n, len(packet))
	}
}

func TestWriteBufferReuseDoesNotCorruptQueuedData(t *testing.T) {
	stackDevice, peerDevice := device.Pipe(1500)
	t.Cleanup(func() { _ = peerDevice.Close() })

	cfg := config.Config{
		StackIP:    "10.0.0.1",
		SendBuf:    8,
		RecvBuf:    64 * 1024,
		MSL:        20 * time.Millisecond,
		MaxRetries: 3,
	}
	stack, err := netstack.New(cfg, stackDevice, telemetry.NewBus(64), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	stack.Start()
	t.Cleanup(func() { _ = stack.Close() })

	type dialResult struct {
		conn *tcp.Conn
		err  error
	}
	dialc := make(chan dialResult, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func() {
		conn, err := stack.Dial(ctx, "10.0.0.2:9000")
		dialc <- dialResult{conn: conn, err: err}
	}()

	syn := readTCPPacket(t, peerDevice)
	if !syn.tcp.Has(tcp.FlagSYN) || len(syn.payload) != 0 {
		t.Fatalf("first packet is not SYN: flags=%v payload=%q", syn.tcp.FlagNames(), syn.payload)
	}
	peerISN := uint32(7000)
	writeTCPPacket(t, peerDevice, syn.ip.Dst, syn.ip.Src, tcp.Header{
		SrcPort: syn.tcp.DstPort,
		DstPort: syn.tcp.SrcPort,
		Seq:     peerISN,
		Ack:     syn.tcp.Seq + 1,
		Flags:   tcp.FlagSYN | tcp.FlagACK,
		Window:  4,
	})

	var conn *tcp.Conn
	select {
	case res := <-dialc:
		if res.err != nil {
			t.Fatalf("dial: %v", res.err)
		}
		conn = res.conn
	case <-time.After(2 * time.Second):
		t.Fatal("dial did not complete")
	}
	ack := readTCPPacket(t, peerDevice)
	if !ack.tcp.Has(tcp.FlagACK) || ack.tcp.Ack != peerISN+1 || len(ack.payload) != 0 {
		t.Fatalf("bad handshake ACK: flags=%v ack=%d payload=%q", ack.tcp.FlagNames(), ack.tcp.Ack, ack.payload)
	}
	t.Cleanup(func() {
		h := tcp.Header{
			SrcPort: syn.tcp.DstPort,
			DstPort: syn.tcp.SrcPort,
			Seq:     peerISN + 1,
			Flags:   tcp.FlagRST,
		}
		segment := tcp.Marshal(&h, syn.ip.Dst, syn.ip.Src, nil)
		packet := ip.Marshal(syn.ip.Dst, syn.ip.Src, ip.ProtoTCP, 64, 2, segment)
		_, _ = peerDevice.Write(packet)
		deadline := time.Now().Add(100 * time.Millisecond)
		for time.Now().Before(deadline) {
			if _, ok := stack.Connection(conn.ID()); !ok {
				return
			}
			time.Sleep(time.Millisecond)
		}
	})

	payload := make([]byte, 8)
	copy(payload, "abcdefgh")
	want := append([]byte(nil), payload...)
	if n, err := conn.Write(payload); err != nil || n != len(payload) {
		t.Fatalf("write: n=%d err=%v", n, err)
	}
	first := readTCPPacket(t, peerDevice)
	if !bytes.Equal(first.payload, want[:4]) {
		t.Fatalf("first segment payload = %q, want %q", first.payload, want[:4])
	}

	copy(payload[4:], "WXYZ")
	writeTCPPacket(t, peerDevice, syn.ip.Dst, syn.ip.Src, tcp.Header{
		SrcPort: syn.tcp.DstPort,
		DstPort: syn.tcp.SrcPort,
		Seq:     peerISN + 1,
		Ack:     syn.tcp.Seq + 1 + uint32(len(first.payload)),
		Flags:   tcp.FlagACK,
		Window:  4,
	})
	var second tcpPacket
	secondSeq := first.tcp.Seq + uint32(len(first.payload))
	for attempts := 0; attempts < 3; attempts++ {
		candidate := readTCPPacket(t, peerDevice)
		if candidate.tcp.Seq == secondSeq {
			second = candidate
			break
		}
	}
	if second.tcp == nil {
		t.Fatalf("did not receive queued segment at sequence %d", secondSeq)
	}
	got := append(append([]byte(nil), first.payload...), second.payload...)
	if !bytes.Equal(got, want) {
		t.Fatalf("peer received %q after caller reused Write buffer; want %q", got, want)
	}
	if len(second.payload) != 4 {
		t.Fatalf("second segment length = %d, want 4", len(second.payload))
	}
}
