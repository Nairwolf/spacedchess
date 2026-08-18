package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"spacedchess/internal/server"
	"spacedchess/internal/store"
)

func listenAddr() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}

func main() {
	store := store.NewStore()

	srv := &http.Server{
		Addr:              listenAddr(),
		Handler:           server.NewServer(store),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
