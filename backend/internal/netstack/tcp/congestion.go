package tcp

type Phase uint8

const (
	PhaseSlowStart Phase = iota
	PhaseAvoidance
	PhaseFastRecovery
)

func (p Phase) String() string {
	switch p {
	case PhaseSlowStart:
		return "slow_start"
	case PhaseAvoidance:
		return "avoidance"
	case PhaseFastRecovery:
		return "fast_recovery"
	default:
		return "unknown"
	}
}

type Congestion struct {
	MSS      uint32
	CWND     uint32
	SSThresh uint32
	Phase    Phase
	Acked    uint32
}

func NewCongestion(mss uint32) *Congestion {
	if mss == 0 {
		mss = 1460
	}
	return &Congestion{
		MSS:      mss,
		CWND:     10 * mss,
		SSThresh: 65535,
		Phase:    PhaseSlowStart,
	}
}

func (c *Congestion) OnACK(acked uint32) {
	if acked == 0 {
		return
	}
	switch c.Phase {
	case PhaseFastRecovery:
		return
	case PhaseSlowStart:
		c.CWND += acked
		if c.CWND >= c.SSThresh {
			c.Phase = PhaseAvoidance
		}
	case PhaseAvoidance:
		c.Acked += acked
		for c.Acked >= c.CWND && c.CWND > 0 {
			c.Acked -= c.CWND
			c.CWND += c.MSS
		}
	}
}

func (c *Congestion) OnTimeout(flight uint32) {
	if flight < 2*c.MSS {
		c.SSThresh = 2 * c.MSS
	} else {
		c.SSThresh = flight / 2
	}
	c.CWND = c.MSS
	c.Phase = PhaseSlowStart
	c.Acked = 0
}

func (c *Congestion) OnFastRetransmit(flight uint32) {
	if flight < 2*c.MSS {
		c.SSThresh = 2 * c.MSS
	} else {
		c.SSThresh = flight / 2
	}
	c.CWND = c.SSThresh + 3*c.MSS
	c.Phase = PhaseFastRecovery
}

func (c *Congestion) OnFastRecoveryACK() {
	c.CWND = c.SSThresh
	c.Phase = PhaseAvoidance
	c.Acked = 0
}

func (c *Congestion) InflateDupACK() {
	if c.Phase == PhaseFastRecovery {
		c.CWND += c.MSS
	}
}

func (c *Congestion) Usable() uint32 {
	if c.CWND == 0 {
		return c.MSS
	}
	return c.CWND
}
