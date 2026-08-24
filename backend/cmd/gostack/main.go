package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gostack/internal/api"
	"gostack/internal/app"
	"gostack/internal/config"
	"gostack/internal/control"
	"gostack/internal/device"
	"gostack/internal/logger"
	"gostack/internal/netstack"
	"gostack/internal/telemetry"
)

func main() {
	cfg := config.Load()
	mode := flag.String("mode", cfg.Device, "tun|mock|native")
	flag.Parse()
	if *mode != "" {
		cfg.Device = *mode
	}
	if cfg.Device == "native" {
		cfg.Device = "tun"
	}
	logger.Init(cfg.LogLevel)
	log := logger.Guard()

	bus := telemetry.NewBus(8192)
	impair := control.NewImpair()
	step := control.NewStep(cfg.StepPPS)

	var (
		dev  device.Device
		peer device.Device
		err  error
	)
	switch cfg.Device {
	case "tun":
		dev, err = device.OpenTUN(cfg.TunName, cfg.MTU)
		if err != nil {
			log.Error("open tun failed, fallback mock", "err", err)
			cfg.Device = "mock"
		} else if err := netstack.ConfigureTUN(dev.Name(), cfg.TunHostIP, 24); err != nil {
			log.Warn("configure tun", "err", err)
		}
	}
	if cfg.Device == "mock" {
		a, b := device.Pipe(cfg.MTU)
		dev, peer = a, b
	}
	if dev == nil {
		log.Error("no device")
		os.Exit(1)
	}

	st, err := netstack.New(cfg, dev, bus, impair, step)
	if err != nil {
		log.Error("stack", "err", err)
		os.Exit(1)
	}
	st.Start()
	if err := app.ServeEcho(st, cfg.ListenAddr); err != nil {
		log.Error("echo", "err", err)
		os.Exit(1)
	}

	rt := &app.Runtime{Cfg: cfg, Stack: st, Bus: bus, Impair: impair, Step: step}
	if peer != nil {
		pcfg := cfg
		pcfg.StackIP = cfg.TunHostIP
		pst, err := netstack.New(pcfg, peer, bus, impair, step)
		if err != nil {
			log.Error("peer stack", "err", err)
			os.Exit(1)
		}
		pst.Start()
		rt.Peer = pst
	}

	srv := api.New(rt)
	if err := srv.Start(); err != nil {
		log.Error("api", "err", err)
		os.Exit(1)
	}
	log.Info("gostack ready", "device", cfg.Device, "listen", cfg.ListenAddr)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	_ = st.Close()
	if rt.Peer != nil {
		_ = rt.Peer.Close()
	}
}
