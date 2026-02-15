package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// enableCORS is a middleware that wraps your handlers.
// It handles the "OPTIONS" preflight request and sets the headers.
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowURL := os.Getenv("CORS_ALLOW")
		if allowURL == "" {
			allowURL = "*" // Fallback for safety, though specific is better
		}

		w.Header().Set("Access-Control-Allow-Origin", allowURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// If the browser is just "asking" for permission (OPTIONS),
		// we stop here and return 204 (No Content).
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Attempt to load .env, but don't crash if it's missing (important for Prod)
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}

	mux := http.NewServeMux()

	// Define your routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World! CORS is configured.")
	})

	// Get port from env or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Wrap the entire mux in the CORS middleware
	fmt.Printf("Server listening on port %s...\n", port)
	err = http.ListenAndServe(":"+port, enableCORS(mux))
	if err != nil {
		log.Fatal(err)
	}
}
