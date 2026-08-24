package tcp

type State uint8

const (
	Closed State = iota
	Listen
	SynSent
	SynRcvd
	Established
	FinWait1
	FinWait2
	CloseWait
	Closing
	LastAck
	TimeWait
)

var stateNames = [...]string{
	Closed:      "CLOSED",
	Listen:      "LISTEN",
	SynSent:     "SYN_SENT",
	SynRcvd:     "SYN_RCVD",
	Established: "ESTABLISHED",
	FinWait1:    "FIN_WAIT_1",
	FinWait2:    "FIN_WAIT_2",
	CloseWait:   "CLOSE_WAIT",
	Closing:     "CLOSING",
	LastAck:     "LAST_ACK",
	TimeWait:    "TIME_WAIT",
}

func (s State) String() string {
	if int(s) < len(stateNames) && stateNames[s] != "" {
		return stateNames[s]
	}
	return "UNKNOWN"
}

func AllStates() []string {
	out := make([]string, 0, 11)
	for i := Closed; i <= TimeWait; i++ {
		out = append(out, i.String())
	}
	return out
}

func ParseState(s string) (State, bool) {
	for i := Closed; i <= TimeWait; i++ {
		if i.String() == s {
			return i, true
		}
	}
	return Closed, false
}

func (s State) CanSendData() bool {
	return s == Established || s == CloseWait
}

func (s State) CanRecvData() bool {
	return s == Established || s == FinWait1 || s == FinWait2
}
