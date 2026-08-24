package logger

import "testing"

func TestInitAndGuard(t *testing.T) {
	Discard()
	Init("debug")
	if Guard() == nil {
		t.Fatal("logger")
	}
	Init("error")
}
