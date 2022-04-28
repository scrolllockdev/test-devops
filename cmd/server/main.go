package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/scrolllockdev/test-devops/internal/server"
)

func main() {
	s := server.Server{}
	httpServer := s.Init()
	ctx, cancel := context.WithCancel(context.Background())
	go httpServer.Run(ctx)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-sig
	s.Shutdown()
	cancel()
}
