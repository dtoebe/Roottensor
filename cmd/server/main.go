package main

import (
	"context"
	"log"

	"github.com/dtoebe/RootTensor/internal/httpserver"
)

func main() {
	srvr := httpserver.NewHTTPServer(":3333")

	if err := srvr.Run(context.Background()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
