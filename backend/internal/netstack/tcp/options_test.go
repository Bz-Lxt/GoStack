package tcp

import "testing"

func TestParseOptions(t *testing.T) {
	raw := append(EncodeMSS(1400), 1, 3, 3, 7, 4, 2, 0)
	o := ParseOptions(raw)
	if !o.HasMSS || o.MSS != 1400 || !o.HasWScale || o.WScale != 7 || !o.SACKPermitted {
		t.Fatalf("%+v", o)
	}
}
