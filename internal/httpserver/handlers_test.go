package httpserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealthz(t *testing.T) {
	svr := setupServer(t)

	t.Run("healthz: response", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		w := httptest.NewRecorder()
		svr.handleHealthz(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("handler returned wrong status code: got: %d; want: %d",
				res.StatusCode, http.StatusOK)
		}

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		got := string(b)
		want := "OK"
		if got != want {
			t.Errorf("handler returned unexpected body: got: %s; want: %s", got, want)
		}
	})

	t.Run("healthz: wrong method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
		w := httptest.NewRecorder()
		svr.handleHealthz(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("handler returned wrong status code: got: %d; want: %d",
				res.StatusCode, http.StatusMethodNotAllowed)
		}
	})
}

func TestHandleIndex(t *testing.T) {
	svr := setupServer(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	svr.handleIndex(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("want: %d; got: %d", http.StatusOK, res.StatusCode)
	}

}

func setupServer(t *testing.T) *HTTPServer {
	t.Helper()

	svr, err := NewHTTPServer("127.0.0.1:0", "../../web/templates", nil)
	if err != nil {
		t.Fatal(err)
	}

	return svr
}
