package main

import (
	"os"
	"os/signal"
	"syscall"
)

// setupShutdownHooks sets up signal handlers for SIGTERM and SIGINT
// and returns the chan to listen on for shutdown.
func setupShutdownHooks() <-chan struct{} {
	shutdown := make(chan struct{}, 1)

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		shutdown <- struct{}{}
	}()

	return shutdown
}
