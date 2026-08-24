package device_test

import (
	"testing"
	"time"

	"gostack/internal/device"
)

func TestPipe(t *testing.T) {
	a, b := device.Pipe(1500)
	go func() { _, _ = a.Write([]byte("ping")) }()
	buf := make([]byte, 16)
	n, err := b.Read(buf)
	if err != nil || string(buf[:n]) != "ping" {
		t.Fatalf("%d %v %q", n, err, buf[:n])
	}
	_ = a.Name()
	_ = a.Close()
	time.Sleep(10 * time.Millisecond)
}
