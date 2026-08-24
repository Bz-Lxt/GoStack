package api

import (
	"net/http"

	"github.com/gorilla/websocket"
	"gostack/internal/logger"
	"gostack/internal/telemetry"
	"gostack/internal/timeutil"
)

func (s *Server) wsTelemetry(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	up := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return allowOrigin(r.Header.Get("Origin"), r.Host, s.rt.Cfg.CORSOrigins)
		},
	}
	conn, err := up.Upgrade(w, r, nil)
	if err != nil {
		logger.Guard().Warn("ws upgrade", "err", err, "origin", origin)
		return
	}
	defer conn.Close()
	id, ch := s.rt.Bus.Subscribe(512)
	defer s.rt.Bus.Unsubscribe(id)
	hello := telemetry.New("conn.open", "", true, map[string]any{
		"local": "ws", "remote": r.RemoteAddr, "role": "listen", "hello": true,
		"time": timeutil.Format(timeutil.Now()),
	})
	if err := conn.WriteJSON(hello); err != nil {
		return
	}
	for ev := range ch {
		if err := conn.WriteJSON(ev); err != nil {
			return
		}
	}
}
