package tcp

import "testing"

func TestSlowStartThenAvoidance(t *testing.T) {
	c := NewCongestion(100)
	c.SSThresh = 400
	c.OnACK(100)
	if c.CWND != 1100 {
		t.Fatalf("cwnd %d", c.CWND)
	}
	c.CWND = 400
	c.OnACK(100)
	if c.Phase != PhaseAvoidance {
		t.Fatalf("phase %s", c.Phase)
	}
}

func TestTimeoutFallback(t *testing.T) {
	c := NewCongestion(100)
	c.CWND = 800
	c.OnTimeout(800)
	if c.CWND != 100 {
		t.Fatalf("cwnd %d", c.CWND)
	}
	if c.SSThresh != 400 {
		t.Fatalf("ssthresh %d", c.SSThresh)
	}
	if c.Phase != PhaseSlowStart {
		t.Fatal(c.Phase)
	}
}

func TestFastRecovery(t *testing.T) {
	c := NewCongestion(100)
	c.OnFastRetransmit(800)
	if c.Phase != PhaseFastRecovery {
		t.Fatal(c.Phase)
	}
	if c.SSThresh != 400 || c.CWND != 700 {
		t.Fatalf("ss=%d cwnd=%d", c.SSThresh, c.CWND)
	}
	c.OnFastRecoveryACK()
	if c.Phase != PhaseAvoidance || c.CWND != 400 {
		t.Fatalf("after %s %d", c.Phase, c.CWND)
	}
}
