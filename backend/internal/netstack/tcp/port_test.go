package tcp

import "testing"

func TestEphemeralUnique(t *testing.T) {
	p := NewPortAlloc()
	a, ok := p.Ephemeral()
	if !ok || a < ephemeralMin {
		t.Fatal(a, ok)
	}
	if !p.Reserve(80) {
		t.Fatal("80")
	}
	if p.Reserve(80) {
		t.Fatal("dup")
	}
	p.Release(80)
	if !p.Reserve(80) {
		t.Fatal("after release")
	}
}
