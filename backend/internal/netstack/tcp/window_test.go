package tcp

import "testing"

func TestSendBufferAckRetrans(t *testing.T) {
	b := NewSendBuffer(1000, 4096)
	if n := b.Write([]byte("abcdef")); n != 6 {
		t.Fatal(n)
	}
	if string(b.Unsent()) != "abcdef" {
		t.Fatal(string(b.Unsent()))
	}
	b.MarkSent(3)
	if string(b.Unsent()) != "def" {
		t.Fatal(string(b.Unsent()))
	}
	if string(b.PeekFromUNA(10)) != "abcdef" {
		t.Fatal("retrans")
	}
	if b.Ack(1003) != 3 {
		t.Fatal("ack")
	}
	if string(b.Unsent()) != "def" {
		t.Fatal(string(b.Unsent()))
	}
}

func TestRecvWindow(t *testing.T) {
	r := NewRecvBuffer(50, 8)
	r.WriteInOrder([]byte("abcd"))
	if r.Window() != 4 {
		t.Fatal(r.Window())
	}
	buf := make([]byte, 2)
	r.Read(buf)
	if string(buf) != "ab" || r.Window() != 6 {
		t.Fatal(string(buf), r.Window())
	}
}
