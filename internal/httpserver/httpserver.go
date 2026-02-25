// Package httpserver
package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HTTPServer struct {
	addr string
}

func NewHTTPServer(addr string) *HTTPServer {
	return &HTTPServer{
		addr: addr,
	}
}

func (s *HTTPServer) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.routes(),
	}

	srvErr := make(chan error, 1)

	go func() {
		log.Printf("server listening on %s", s.addr)
		if err := srv.ListenAndServe(); err != nil {
			srvErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-srvErr:
		return fmt.Errorf("server failed to start: %v", err)
	case <-stop:
		log.Println("shutdown signal received")
	case <-ctx.Done():
		log.Println("context cancelled, shutting down")
	}

	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("server shut down")
	return nil
}
