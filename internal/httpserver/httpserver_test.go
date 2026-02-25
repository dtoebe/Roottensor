package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthzHandler(t *testing.T) {
	svr := NewHTTPServer("127.0.0.1:0")

	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	hdlr := http.HandlerFunc(svr.handleHealthz)

	hdlr.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got: %d; want: %d", status, http.StatusOK)
	}

	want := "OK"
	if got := rr.Body.String(); got != want {
		t.Errorf("handler returned unexpected body: got: %s; want: %s", got, want)
	}
}

func TestServerLifecycle(t *testing.T) {
	addr := ":3333"
	srv := NewHTTPServer(addr)

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
