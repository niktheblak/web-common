package graceful

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

type Shutdown struct {
	Server          Server
	ShutdownTimeout time.Duration
	Signals         []os.Signal
}

func (g *Shutdown) Serve(ctx context.Context) error {
	if g.Server == nil {
		return fmt.Errorf("server cannot be nil")
	}
	if len(g.Signals) > 0 {
		var stop context.CancelFunc
		ctx, stop = signal.NotifyContext(ctx, g.Signals...)
		defer stop()
	}
	// serveCtx is canceled by the caller's context, by one of the given signals or by the server
	// itself no longer serving, whichever happens first.
	serveCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errChan := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		// ListenAndServe returning at all means the server has stopped serving, so shut down
		// even if it reported no error.
		defer cancel()
		if err := g.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()
	go func() {
		defer wg.Done()
		<-serveCtx.Done()
		// The shutdown context must not be derived from serveCtx, which is always canceled by
		// the time we get here; that would abort the shutdown before it could drain anything.
		shutdownCtx := context.Background()
		if g.ShutdownTimeout > 0 {
			var shutdownCancel context.CancelFunc
			shutdownCtx, shutdownCancel = context.WithTimeout(shutdownCtx, g.ShutdownTimeout)
			defer shutdownCancel()
		}
		if err := g.Server.Shutdown(shutdownCtx); err != nil {
			errChan <- err
		}
	}()
	wg.Wait()
	close(errChan)
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
