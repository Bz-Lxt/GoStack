package telemetry

import (
	"encoding/json"

	"gostack/internal/timeutil"
)

type Event struct {
	V       int            `json:"v"`
	TS      string         `json:"ts"`
	Type    string         `json:"type"`
	ConnID  string         `json:"conn_id"`
	Precise bool           `json:"precise"`
	Payload map[string]any `json:"payload"`
}

func New(typ, connID string, precise bool, payload map[string]any) Event {
	if payload == nil {
		payload = map[string]any{}
	}
	return Event{
		V: 1, TS: timeutil.FormatRFC3339(timeutil.Now()),
		Type: typ, ConnID: connID, Precise: precise, Payload: payload,
	}
}

func (e Event) Bytes() []byte {
	b, _ := json.Marshal(e)
	return b
}
