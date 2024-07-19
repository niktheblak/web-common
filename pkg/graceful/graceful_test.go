package graceful

import (
	"context"
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
	closed bool
}

func (s *testServer) ListenAndServe() error {
	s.server.Start()
	return nil
}

func (s *testServer) Shutdown(ctx context.Context) error {
	s.server.Close()
	s.closed = true
	return nil
}

func TestServe(t *testing.T) {
	server := &testServer{
		server: httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Test server")
		})),
	}
	g := Graceful{
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
