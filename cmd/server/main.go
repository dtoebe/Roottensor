package main

import (
	"context"
	"log"

	"github.com/dtoebe/RootTensor/internal/httpserver"
)

func main() {
	srvr, err := httpserver.NewHTTPServer(":3333", "web/templates")
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	if err := srvr.Run(context.Background()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
