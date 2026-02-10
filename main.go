package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"stock-analysis/internal/api"
	"time"
)

func main() {
	port := "8000"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	handler := api.Handler()
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	fmt.Printf("Starting Fear & Greed Server on http://localhost:%s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
