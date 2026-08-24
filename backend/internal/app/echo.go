package app

import (
	"io"
	"sync"

	"gostack/internal/logger"
	"gostack/internal/netstack"
)

func ServeEcho(st *netstack.Stack, addr string) error {
	ln, err := st.Listen(addr)
	if err != nil {
		return err
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer c.Close()
				_, _ = io.Copy(c, c)
			}()
		}
	}()
	logger.Guard().Info("echo listen", "addr", addr)
	return nil
}

type onceCloser struct {
	once sync.Once
	fn   func()
}

func (o *onceCloser) Close() { o.once.Do(o.fn) }
