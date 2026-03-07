package httpserver

import (
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dtoebe/RootTensor/internal/templates"
)

func (s *HTTPServer) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handlePage("Home", templates.HomePage()))
	mux.HandleFunc("GET /about", s.handlePage("About", templates.AboutPage()))
	mux.HandleFunc("GET /settings", s.handlePage("Settings", templates.SettingsPage()))
	mux.HandleFunc("GET /healthz", s.handleHealthz)

	mux.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("web/static"))))

	return mux
}

func (s *HTTPServer) handlePage(title string, content templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.Layout(title, r.URL.Path, content).Render(r.Context(), w); err != nil {
			http.Error(w, "render error", http.StatusInternalServerError)
			log.Printf("layout render error: %v", err)
		}
	}
}

func (s *HTTPServer) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "plain/text; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf("error writing response: %s; error: %v", r.URL, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
