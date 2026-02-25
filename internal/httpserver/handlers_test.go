package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthzHandler(t *testing.T) {
	svr, err := NewHTTPServer("127.0.0.1:0", "../../web/templates")
	if err != nil {
		t.Fatal(err)
	}

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
