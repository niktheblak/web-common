package graceful

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testServer struct {
	server *httptest.Server
	done   chan struct{}
	closed bool
}

func newTestServer(t *testing.T) *testServer {
	return &testServer{
		server: httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := fmt.Fprintln(w, "Test server"); err != nil {
				t.Error(err)
			}
		})),
		done: make(chan struct{}),
	}
}

// ListenAndServe blocks until Shutdown is called, like http.Server does.
func (s *testServer) ListenAndServe() error {
	s.server.Start()
	<-s.done
	return http.ErrServerClosed
}

func (s *testServer) Shutdown(ctx context.Context) error {
	s.server.Close()
	s.closed = true
	close(s.done)
	return nil
}

// failingServer fails to bind, the way http.Server does when its port is already taken.
type failingServer struct {
	err error
}

func (s *failingServer) ListenAndServe() error {
	return s.err
}

func (s *failingServer) Shutdown(ctx context.Context) error {
	return nil
}

func TestServe(t *testing.T) {
	server := newTestServer(t)
	g := Shutdown{
		Server:          server,
		ShutdownTimeout: 1 * time.Second,
		Signals:         []os.Signal{syscall.SIGUSR1},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := g.Serve(ctx)
		wg.Done()
		assert.NoError(t, err)
	}()
	time.Sleep(100 * time.Millisecond)
	err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	require.NoError(t, err)
	wg.Wait()
	assert.True(t, server.closed)
}

func TestServeReturnsListenAndServeErrorWithoutSignals(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("listen tcp :8080: bind: address already in use")
	g := Shutdown{
		Server:          &failingServer{err: wantErr},
		ShutdownTimeout: 1 * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Without signals to listen for, Serve used to block forever here instead of reporting
	// that the server never came up.
	err := g.Serve(ctx)
	assert.ErrorIs(t, err, wantErr)
	assert.NoError(t, ctx.Err(), "Serve should have returned well before the context deadline")
}

func TestServeReturnsWhenCallerContextIsCanceled(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	g := Shutdown{
		Server:          server,
		ShutdownTimeout: 1 * time.Second,
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	assert.NoError(t, g.Serve(ctx))
	assert.True(t, server.closed)
}
