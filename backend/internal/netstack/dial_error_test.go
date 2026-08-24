package netstack_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gostack/internal/netstack/tcp"
)

func TestDialReportsPeerRefusal(t *testing.T) {
	client, _ := pair(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := client.Dial(ctx, "10.0.0.2:9101")
	gotConn := conn != nil
	if conn != nil {
		_ = conn.Close()
	}
	if gotConn || !errors.Is(err, tcp.ErrRefused) {
		t.Fatalf("dial after peer refusal = (connection returned: %t, error: %v), want (false, tcp.ErrRefused)", gotConn, err)
	}
}
