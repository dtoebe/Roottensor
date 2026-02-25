package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestServerLifecycle(t *testing.T) {
	addr := ":3333"
	srv, err := NewHTTPServer(addr, "../../web/templates")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost%s/healthz", addr))
	if err != nil {
		t.Fatalf("failed to make request to %s: %v", addr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want: %d; got: %d", http.StatusOK, resp.StatusCode)
	}

	cancel()

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("server exited with error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server failed to shutdown in time")
	}
}
