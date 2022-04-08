package main

import (
	"context"

	"github.com/scrolllockdev/test-devops/internal/server"
)

func main() {
	s := server.NewServer()

	ctx, cancel := context.WithCancel(context.Background())
	s.Run(ctx)

	s.Shutdown() //blocking function
	cancel()

}
