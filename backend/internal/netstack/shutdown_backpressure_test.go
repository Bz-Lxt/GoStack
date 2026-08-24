package netstack_test

import (
	"testing"
	"time"

	"gostack/internal/config"
	"gostack/internal/control"
	"gostack/internal/device"
	"gostack/internal/netstack"
	"gostack/internal/telemetry"
)

func TestStackCloseReleasesBackpressuredDeviceWrite(t *testing.T) {
	dev, peer := device.Pipe(1500)
	t.Cleanup(func() { _ = peer.Close() })
	st, err := netstack.New(
		config.Config{StackIP: "10.0.0.2"},
		dev,
		telemetry.NewBus(16),
		control.NewImpair(),
		control.NewStep(20),
	)
	if err != nil {
		t.Fatal(err)
	}
	st.Start()

	var blockedWrite <-chan error
	for attempts := 0; attempts < 4096; attempts++ {
		result := make(chan error, 1)
		started := make(chan struct{})
		go func() {
			close(started)
			_, err := dev.Write([]byte{1})
			result <- err
		}()
		<-started

		select {
		case err := <-result:
			if err != nil {
				t.Fatalf("fill device output: %v", err)
			}
		case <-time.After(100 * time.Millisecond):
			blockedWrite = result
		}
		if blockedWrite != nil {
			break
		}
	}
	if blockedWrite == nil {
		_ = st.Close()
		t.Fatal("device output never became backpressured")
	}
	time.Sleep(20 * time.Millisecond)

	closeDone := make(chan error, 1)
	go func() { closeDone <- st.Close() }()
	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("close stack: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stack close blocked behind a backpressured device write")
	}

	select {
	case err := <-blockedWrite:
		if err == nil {
			t.Fatal("backpressured write succeeded after device close")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("backpressured write was not released by device close")
	}
}
