package signals

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func SetupSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		cancel()
		os.Exit(1)
	}()
	return ctx
}
