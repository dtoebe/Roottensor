package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealthz(t *testing.T) {
	svr := setupServer(t)

	req, err := http.NewRequest(http.MethodGet, "/healthz", nil)
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

	svr, err := NewHTTPServer("127.0.0.1:0", "../../web/templates")
	if err != nil {
		t.Fatal(err)
	}

	return svr
}
