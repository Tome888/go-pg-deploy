package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// App holds our application dependencies, like the database pool
type App struct {
	DB *pgxpool.Pool
}

func main() {
	// 1. Load envs (local only)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// 2. Connect to the Database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set in environment variables")
	}

	// Create a connection pool (better for performance than a single connection)
	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	defer dbPool.Close()

	// Verify the connection works
	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("Could not ping database: %v\n", err)
	}
	fmt.Println("Successfully connected to PostgreSQL!")

	// 3. Initialize App and Routes
	app := &App{DB: dbPool}
	mux := http.NewServeMux()

	mux.HandleFunc("/", app.HomeHandler)
	mux.HandleFunc("/health", app.HealthHandler)

	// 4. Start the Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default for Render
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      enableCORS(mux),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("Server listening on port %s...\n", port)
	log.Fatal(server.ListenAndServe())
}

// HomeHandler demonstrates a simple query
func (a *App) HomeHandler(w http.ResponseWriter, r *http.Request) {
	var version string
	err := a.DB.QueryRow(context.Background(), "SELECT version()").Scan(&version)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Hello! Database version: %s", version)
}

func (a *App) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// enableCORS middleware
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowURL := os.Getenv("CORS_ALLOW")
		if allowURL == "" {
			allowURL = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", allowURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
