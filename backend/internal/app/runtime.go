package app

import (
	"gostack/internal/config"
	"gostack/internal/control"
	"gostack/internal/netstack"
	"gostack/internal/telemetry"
)

type Runtime struct {
	Cfg    config.Config
	Stack  *netstack.Stack
	Peer   *netstack.Stack
	Bus    *telemetry.Bus
	Impair *control.Impair
	Step   *control.Step
}
