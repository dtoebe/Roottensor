package httpserver

import "net/http"

func (s *HTTPServer) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/healthz", s.handleHealthz)

	mux.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("web/static"))))

	return mux
}

func (s *HTTPServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if err := s.templates.index.ExecuteTemplate(w, "layout.html", nil); err != nil {
		http.Error(w, "error loading page", http.StatusInternalServerError)
	}
}

func (s *HTTPServer) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "plain/text; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
