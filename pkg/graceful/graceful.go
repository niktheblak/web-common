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
	errChan := make(chan error, 2)
	var cancel context.CancelFunc = nil
	if len(g.Signals) > 0 {
		ctx, cancel = signal.NotifyContext(ctx, g.Signals...)
		defer cancel()
	}
	go func() {
		if err := g.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
			if cancel != nil {
				cancel()
			}
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		var shutdownCtx context.Context
		if g.ShutdownTimeout > 0 {
			shutdownCtx = context.Background()
			var shutdownCancel context.CancelFunc
			shutdownCtx, shutdownCancel = context.WithTimeout(ctx, g.ShutdownTimeout)
			defer shutdownCancel()
		} else {
			shutdownCtx = ctx
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
