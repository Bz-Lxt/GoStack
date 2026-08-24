package api

import (
	"context"
	"net"
	"net/http"
	"time"

	"gostack/internal/app"
	"gostack/internal/logger"
)

type Server struct {
	rt  *app.Runtime
	http *http.Server
}

func New(rt *app.Runtime) *Server {
	s := &Server{rt: rt}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", s.health)
	mux.HandleFunc("GET /api/v1/connections", s.connections)
	mux.HandleFunc("GET /api/v1/connections/{id}", s.connection)
	mux.HandleFunc("GET /api/v1/stack/stats", s.stats)
	mux.HandleFunc("POST /api/v1/control/step-mode", s.stepMode)
	mux.HandleFunc("POST /api/v1/control/impair", s.impair)
	mux.HandleFunc("POST /api/v1/traffic/start", s.traffic)
	mux.HandleFunc("GET /ws/telemetry", s.wsTelemetry)
	s.http = &http.Server{
		Addr:              rt.Cfg.HTTPAddr,
		Handler:           withCORS(mux, rt.Cfg.CORSOrigins),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.rt.Cfg.HTTPAddr)
	if err != nil {
		return err
	}
	logger.Guard().Info("http listen", "addr", ln.Addr().String())
	go func() {
		if err := s.http.Serve(ln); err != nil && err != http.ErrServerClosed {
			logger.Guard().Error("http", "err", err)
		}
	}()
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
