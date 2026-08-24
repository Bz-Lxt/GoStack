package netstack_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"testing"
	"time"

	"gostack/internal/config"
	"gostack/internal/control"
	"gostack/internal/device"
	"gostack/internal/logger"
	"gostack/internal/netstack"
	"gostack/internal/netstack/tcp"
	"gostack/internal/telemetry"
)

func init() { logger.Discard() }

func pair(t *testing.T) (*netstack.Stack, *netstack.Stack) {
	t.Helper()
	a, b := device.Pipe(1500)
	cfgA := config.Config{StackIP: "10.0.0.1", ListenAddr: "10.0.0.1:1", SendBuf: 64 * 1024, RecvBuf: 64 * 1024, MSL: 200 * time.Millisecond, DemoRTOMin: 200 * time.Millisecond, MaxRetries: 8, MTU: 1500}
	cfgB := cfgA
	cfgB.StackIP = "10.0.0.2"
	sa, err := netstack.New(cfgA, a, telemetry.NewBus(1024), control.NewImpair(), control.NewStep(20))
	if err != nil {
		t.Fatal(err)
	}
	sb, err := netstack.New(cfgB, b, telemetry.NewBus(1024), control.NewImpair(), control.NewStep(20))
	if err != nil {
		t.Fatal(err)
	}
	sa.Start()
	sb.Start()
	t.Cleanup(func() { _ = sa.Close(); _ = sb.Close() })
	return sa, sb
}

func TestHandshakeEchoClose(t *testing.T) {
	sa, sb := pair(t)
	ln, err := sb.Listen("10.0.0.2:9000")
	if err != nil {
		t.Fatal(err)
	}
	errc := make(chan error, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			errc <- err
			return
		}
		_, _ = io.Copy(c, c)
		_ = c.Close()
		errc <- nil
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := sa.Dial(ctx, "10.0.0.2:9000")
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("hello-gostack")
	if _, err := c.Write(msg); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, len(msg))
	if _, err := io.ReadFull(c, buf); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, msg) {
		t.Fatalf("echo %q", buf)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-errc:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server hang")
	}
	seen := map[string]bool{}
	for _, s := range append(sa.Connections(), sb.Connections()...) {
		seen[s.State] = true
	}
	_ = seen
}

func TestBulkSHA256(t *testing.T) {
	sa, sb := pair(t)
	ln, err := sb.Listen("10.0.0.2:9000")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		_, _ = io.Copy(c, c)
		_ = c.Close()
	}()
	c, err := sa.Dial(context.Background(), "10.0.0.2:9000")
	if err != nil {
		t.Fatal(err)
	}
	payload := make([]byte, 64*1024)
	_, _ = rand.Read(payload)
	want := sha256.Sum256(payload)
	if _, err := c.Write(payload); err != nil {
		t.Fatal(err)
	}
	got := make([]byte, 0, len(payload))
	buf := make([]byte, 4096)
	deadline := time.Now().Add(8 * time.Second)
	_ = c.SetReadDeadline(deadline)
	for len(got) < len(payload) {
		n, err := c.Read(buf)
		if n > 0 {
			got = append(got, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	if sha256.Sum256(got) != want {
		t.Fatalf("len=%d", len(got))
	}
}

func TestStateNamesCovered(t *testing.T) {
	if len(tcp.AllStates()) != 11 {
		t.Fatal(tcp.AllStates())
	}
}
