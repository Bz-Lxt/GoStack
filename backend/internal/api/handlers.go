package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gostack/internal/app"
	"gostack/internal/netstack/tcp"
	"gostack/internal/timeutil"
)

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeOK(w, map[string]any{
		"status": "ok",
		"time":   timeutil.Format(timeutil.Now()),
		"device": s.rt.Stack.DeviceName(),
		"tz":     "Asia/Shanghai",
		"mode":   s.rt.Cfg.Device,
	})
}

func (s *Server) connections(w http.ResponseWriter, r *http.Request) {
	list := s.rt.Stack.Connections()
	if list == nil {
		list = []tcp.Snapshot{}
	}
	writeOK(w, list)
}

func (s *Server) connection(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = strings.TrimPrefix(r.URL.Path, "/api/v1/connections/")
	}
	if id == "" {
		writeErr(w, 400, "bad_request", "missing id")
		return
	}
	snap, ok := s.rt.Stack.Connection(id)
	if !ok {
		writeErr(w, 404, "not_found", "connection not found")
		return
	}
	writeOK(w, snap)
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	writeOK(w, s.rt.Stack.Stats().Snapshot())
}

func (s *Server) stepMode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enabled *bool `json:"enabled"`
		PPS     int   `json:"pps"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "bad_request", "invalid json")
		return
	}
	en := false
	if body.Enabled != nil {
		en = *body.Enabled
	}
	s.rt.Step.Set(en, body.PPS)
	writeOK(w, s.rt.Step.Snapshot())
}

func (s *Server) impair(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Loss     float64 `json:"loss"`
		DelayMS  int     `json:"delay_ms"`
		JitterMS int     `json:"jitter_ms"`
		Reorder  float64 `json:"reorder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "bad_request", "invalid json")
		return
	}
	if body.Loss < 0 || body.Loss > 1 || body.Reorder < 0 || body.Reorder > 1 {
		writeErr(w, 400, "bad_request", "loss/reorder must be 0..1")
		return
	}
	s.rt.Impair.Configure(body.Loss, time.Duration(body.DelayMS)*time.Millisecond, time.Duration(body.JitterMS)*time.Millisecond, body.Reorder)
	writeOK(w, s.rt.Impair.Snapshot())
}

func (s *Server) traffic(w http.ResponseWriter, r *http.Request) {
	var req app.TrafficReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.ContentLength > 0 {
		writeErr(w, 400, "bad_request", "invalid json")
		return
	}
	st := s.rt.Stack
	if s.rt.Peer != nil {
		st = s.rt.Peer
	}
	peer := s.rt.Cfg.ListenAddr
	data, err := app.StartTraffic(r.Context(), st, s.rt.Cfg.Device, peer, req)
	if err != nil {
		writeErr(w, 500, "traffic", err.Error())
		return
	}
	writeOK(w, data)
}
