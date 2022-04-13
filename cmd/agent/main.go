package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/scrolllockdev/test-devops/internal/agent"
	"github.com/scrolllockdev/test-devops/internal/agent/config"
)

func main() {

	cfg := config.Config{}
	if err := cfg.ReadConfig(); err != nil {
		panic(err)
	}

	a := agent.Agent{}
	agent := a.Init(cfg)

	context, cancel := context.WithCancel(context.Background())
	go agent.Run(context)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-sig

	cancel()

}
