package main

import (
	"ip2country-service/api"
	"ip2country-service/config"
	"ip2country-service/internal/database"
	"ip2country-service/internal/rate_limiter"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	log.Println("Loading configuration...")
	cfg := config.LoadConfig()
	log.Println("Configuration loaded successfully.")

	// Initialize the database (MongoDB, JSON, or other)
	log.Println("Initializing the database...")
	db, err := database.NewIPDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully.")

	// Initialize the router
	log.Println("Initializing the router...")
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// Log all requests for debugging
	apiRouter.Use(loggingMiddleware)

	// Middleware (rate limiting)
	log.Println("Initializing rate limiter...")
	rl, err := rate_limiter.NewRateLimiter(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize rate limiter: %v", err)
	}
	apiRouter.Use(rl.Limit)
	log.Println("Rate limiter initialized successfully.")

	// Register API handlers
	log.Println("Registering API handlers...")
	api.RegisterHandlers(apiRouter, db, cfg)
	log.Println("API handlers registered successfully.")

	// Start the server
	log.Printf("Server is running on port %s...", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// loggingMiddleware logs incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
