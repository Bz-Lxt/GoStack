package netstack_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gostack/internal/config"
	"gostack/internal/control"
	"gostack/internal/device"
	"gostack/internal/netstack"
	"gostack/internal/netstack/tcp"
	"gostack/internal/telemetry"
)

func TestDialCancellationAbortsPendingHandshake(t *testing.T) {
	cfg := config.Config{
		StackIP:   "10.0.0.1",
		SendBuf:   64 * 1024,
		RecvBuf:   64 * 1024,
		MSL:       200 * time.Millisecond,
		DemoRTOMin: 200 * time.Millisecond,
		MaxRetries: 8,
		MTU:        1500,
	}
	st, err := netstack.New(
		cfg,
		device.NewMock("silent-peer", 1500),
		telemetry.NewBus(16),
		control.NewImpair(),
		control.NewStep(20),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	type dialResult struct {
		conn *tcp.Conn
		err  error
	}
	result := make(chan dialResult, 1)
	go func() {
		conn, err := st.Dial(ctx, "10.0.0.2:9000")
		result <- dialResult{conn: conn, err: err}
	}()

	waitUntil := time.Now().Add(time.Second)
	for st.Stats().Snapshot()["conns"] == 0 && time.Now().Before(waitUntil) {
		time.Sleep(time.Millisecond)
	}
	if st.Stats().Snapshot()["conns"] == 0 {
		cancel()
		t.Fatal("Dial did not enter a pending handshake")
	}

	cancel()
	select {
	case got := <-result:
		if got.conn != nil {
			_ = got.conn.Close()
			t.Fatal("canceled Dial returned a connection")
		}
		if !errors.Is(got.err, context.Canceled) {
			t.Fatalf("Dial error = %v, want context.Canceled", got.err)
		}
	case <-time.After(500 * time.Millisecond):
		pending := st.Connections()
		_ = st.Close()
		select {
		case got := <-result:
			if got.conn != nil {
				_ = got.conn.Close()
			}
		case <-time.After(time.Second):
		}
		t.Fatalf("Dial remained blocked after cancellation; connections: %+v", pending)
	}

	if conns := st.Connections(); len(conns) != 0 {
		t.Fatalf("canceled Dial retained connections: %+v", conns)
	}
}
