package main

import (
	"log"
	"net/http"
	"os"

	"messaging-app/app"
)

func main() {
	// Initialize application
	application := app.NewApp()

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, application.Handler()); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
