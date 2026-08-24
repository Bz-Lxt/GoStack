package tcp

import "testing"

func TestElevenStates(t *testing.T) {
	all := AllStates()
	if len(all) != 11 {
		t.Fatalf("want 11 got %d %v", len(all), all)
	}
	for _, s := range all {
		if _, ok := ParseState(s); !ok {
			t.Fatalf("parse %s", s)
		}
	}
}
