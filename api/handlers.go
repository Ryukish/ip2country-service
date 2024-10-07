package api

import (
	v1 "ip2country-service/api/v1"
	"ip2country-service/config"
	"ip2country-service/internal/database"
	"net/http"

	"github.com/gorilla/mux"
)

// HealthCheckHandler handles the health check requests
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service is up and running"))
}

// RegisterHandlers registers all the API routes and their corresponding handlers
func RegisterHandlers(router *mux.Router, db database.IPDatabase, cfg *config.Config) {
	// Create the handler for IP lookups
	ipHandler := v1.NewIPHandler(db, cfg)

	// Register API route for getting IP location
	router.HandleFunc("/find-country", ipHandler.GetLocation).Methods(http.MethodGet)

	// Register health check endpoint
	router.HandleFunc("/health", HealthCheckHandler).Methods(http.MethodGet)
}
