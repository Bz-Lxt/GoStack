package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"net"
	"time"

	"gostack/internal/logger"
	"gostack/internal/netstack"
)

type TrafficReq struct {
	Bytes    int    `json:"bytes"`
	RatePPS  int    `json:"rate_pps"`
	Scenario string `json:"scenario"`
}

func StartTraffic(ctx context.Context, st *netstack.Stack, device string, peer string, req TrafficReq) (map[string]any, error) {
	if req.Bytes <= 0 {
		req.Bytes = 64 * 1024
	}
	if req.Bytes > 8*1024*1024 {
		req.Bytes = 8 * 1024 * 1024
	}
	payload := make([]byte, req.Bytes)
	_, _ = rand.Read(payload)
	sum := sha256.Sum256(payload)

	switch device {
	case "tun":
		d, err := net.DialTimeout("tcp", peer, 5*time.Second)
		if err != nil {
			return nil, fmt.Errorf("kernel dial: %w", err)
		}
		defer d.Close()
		if _, err := d.Write(payload); err != nil {
			return nil, err
		}
		_ = d.(*net.TCPConn).CloseWrite()
		got, err := io.ReadAll(d)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"bytes": len(got), "sha256": fmt.Sprintf("%x", sha256.Sum256(got)),
			"match": sha256.Sum256(got) == sum, "via": "kernel",
		}, nil
	default:
		c, err := st.Dial(ctx, peer)
		if err != nil {
			return nil, fmt.Errorf("stack dial: %w", err)
		}
		defer c.Close()
		if _, err := c.Write(payload); err != nil {
			return nil, err
		}
		got := make([]byte, 0, len(payload))
		buf := make([]byte, 4096)
		deadline := time.Now().Add(10 * time.Second)
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
		logger.Guard().Info("mock traffic done", "n", len(got))
		return map[string]any{
			"bytes": len(got), "sha256": fmt.Sprintf("%x", sha256.Sum256(got)),
			"match": sha256.Sum256(got) == sum, "via": "stack",
		}, nil
	}
}
