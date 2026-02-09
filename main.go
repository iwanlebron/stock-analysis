package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"stock-analysis/internal/api"
)

func main() {
	port := "8000"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	handler := api.Handler()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	fmt.Printf("Starting Fear & Greed Server on http://localhost:%s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
