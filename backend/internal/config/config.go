package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Device      string
	TunName     string
	StackIP     string
	TunHostIP   string
	ListenAddr  string
	HTTPAddr    string
	LogLevel    string
	CORSOrigins []string
	MSL         time.Duration
	DemoRTOMin  time.Duration
	MaxRetries  int
	RecvBuf     int
	SendBuf     int
	MTU         int
	StepPPS     int
}

func Load() Config {
	c := Config{
		Device:     env("DEVICE", "mock"),
		TunName:    env("TUN_NAME", "tun0"),
		StackIP:    env("STACK_IP", "10.0.0.2"),
		TunHostIP:  env("TUN_HOST_IP", "10.0.0.1"),
		ListenAddr: env("LISTEN_ADDR", "10.0.0.2:9000"),
		HTTPAddr:   env("HTTP_ADDR", ":8080"),
		LogLevel:   env("LOG_LEVEL", "info"),
		MSL:        time.Duration(envInt("MSL_SEC", 30)) * time.Second,
		DemoRTOMin: time.Duration(envInt("RTO_MIN_MS", 200)) * time.Millisecond,
		MaxRetries: envInt("MAX_RETRIES", 5),
		RecvBuf:    envInt("RECV_BUF", 64*1024),
		SendBuf:    envInt("SEND_BUF", 64*1024),
		MTU:        envInt("MTU", 1500),
		StepPPS:    envInt("STEP_PPS", 20),
	}
	raw := env("CORS_ORIGINS", "http://127.0.0.1:28471,http://localhost:28471")
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			c.CORSOrigins = append(c.CORSOrigins, p)
		}
	}
	return c
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
