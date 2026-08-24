package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	mu sync.Mutex
	L  *slog.Logger
)

func Init(level string) {
	mu.Lock()
	defer mu.Unlock()
	L = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: parse(level)}))
	slog.SetDefault(L)
}

func parse(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Guard() *slog.Logger {
	if L == nil {
		Init("info")
	}
	return L
}

func Discard() {
	mu.Lock()
	defer mu.Unlock()
	L = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	slog.SetDefault(L)
}
