package main

import (
	"context"
	"log"

	"github.com/dtoebe/RootTensor/internal/httpserver"
	"github.com/dtoebe/RootTensor/internal/store"
)

func main() {
	db, err := store.NewSQLiteDB("roottensor.db")
	srvr, err := httpserver.NewHTTPServer(":3333", "web/templates", db)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	if err := srvr.Run(context.Background()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
