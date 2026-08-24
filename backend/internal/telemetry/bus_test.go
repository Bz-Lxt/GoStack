package telemetry_test

import (
	"testing"
	"time"

	"gostack/internal/telemetry"
)

func TestBusDropOldest(t *testing.T) {
	b := telemetry.NewBus(4)
	for i := 0; i < 8; i++ {
		b.Publish(telemetry.New("x", "", true, map[string]any{"i": i}))
	}
	if b.Dropped() == 0 {
		t.Fatal("expected drops")
	}
}

func TestSubscribeNonBlocking(t *testing.T) {
	b := telemetry.NewBus(8)
	id, ch := b.Subscribe(1)
	defer b.Unsubscribe(id)
	b.Publish(telemetry.New("a", "", true, nil))
	b.Publish(telemetry.New("b", "", true, nil))
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
