package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gostack/internal/app"
	"gostack/internal/config"
	"gostack/internal/control"
	"gostack/internal/device"
	"gostack/internal/logger"
	"gostack/internal/netstack"
	"gostack/internal/telemetry"
)

func TestHealth(t *testing.T) {
	logger.Discard()
	a, b := device.Pipe(1500)
	_ = b.Close()
	cfg := config.Config{StackIP: "10.0.0.2", HTTPAddr: ":0", Device: "mock", CORSOrigins: []string{}}
	st, err := netstack.New(cfg, a, telemetry.NewBus(16), control.NewImpair(), control.NewStep(20))
	if err != nil {
		t.Fatal(err)
	}
	s := New(&app.Runtime{Cfg: cfg, Stack: st, Bus: telemetry.NewBus(8), Impair: control.NewImpair(), Step: control.NewStep(20)})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	s.http.Handler.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	var env map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	if env["ok"] != true {
		t.Fatal(env)
	}
}
