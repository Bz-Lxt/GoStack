package ip_test

import (
	"testing"

	"gostack/internal/netstack/ip"
)

func TestEchoReply(t *testing.T) {
	req := (&ip.ICMPEcho{Type: ip.ICMPEchoRequest, ID: 7, Seq: 3, Data: []byte("ping")}).Marshal()
	m, ok := ip.ParseICMPEcho(req)
	if !ok || m.Type != ip.ICMPEchoRequest {
		t.Fatal(m, ok)
	}
	rep := ip.EchoReply(req)
	got, ok := ip.ParseICMPEcho(rep)
	if !ok || got.Type != ip.ICMPEchoReply || string(got.Data) != "ping" {
		t.Fatalf("%+v", got)
	}
}
