package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/scrolllockdev/test-devops/internal/server"
	"github.com/scrolllockdev/test-devops/internal/server/config"
)

func main() {

	fmt.Println("start server")

	cfg := config.Config{}
	if err := cfg.ReadConfig(); err != nil {
		panic(err)
	}

	s := server.Server{}
	httpServer := s.Init(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	go httpServer.Run(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-sig

	s.Shutdown()

	fmt.Println("Stop server")

	cancel()

}
