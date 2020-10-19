package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// NamedCloser wraps io.Closer with Name() method
type NamedCloser interface {
	io.Closer
	Name() string
}

func gracefulShutdown(ctx context.Context, components ...NamedCloser) <-chan struct{} {
	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, os.Interrupt, syscall.SIGTERM)
	end := make(chan struct{})
	stop := func() {
		for _, c := range components {
			if err := c.Close(); err != nil {
				log.WithError(err).WithField("component", c.Name()).Warn("Unable to close the component")
			}
		}
		close(end)
	}

	go func() {
		for {
			select {
			case sig := <-gracefulStop:
				log.Warnf("Stopping the app due to a caught sig %+v", sig)
				stop()
			case <-ctx.Done():
				log.Warn("Stopping the app due to a canceled context")
				stop()
			}
		}
	}()

	return end
}
