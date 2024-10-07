package v1

import (
	"encoding/json"
	"ip2country-service/config"
	"ip2country-service/internal/database"
	"ip2country-service/internal/models"
	"ip2country-service/pkg/utils"
	"log"
	"net/http"
	"strings"
)

type IPHandler struct {
	db     database.IPDatabase
	config *config.Config
}

func NewIPHandler(db database.IPDatabase, cfg *config.Config) *IPHandler {
	return &IPHandler{db: db, config: cfg}
}

func (h *IPHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	fields := r.URL.Query().Get("fields")

	log.Printf("Received request for IP: %s with fields: %s", ip, fields)

	// Validate the IP
	if !utils.ValidateIP(ip) {
		log.Printf("Invalid IP address: %s", ip)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid IP address")
		return
	}

	// Query the database for the IP location
	log.Printf("Querying database for IP: %s", ip)
	loc, err := h.db.Find(ip)
	if err != nil {
		if err.Error() == "not found" { // You can improve this with a custom error type
			log.Printf("IP not found in the database: %s", ip)
			utils.RespondWithError(w, http.StatusNotFound, "IP not found")
			return
		}
		log.Printf("Error querying MongoDB for IP %s: %v", ip, err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	log.Printf("IP found: %+v", loc)

	// Build the response
	log.Printf("Building response for IP: %s", ip)
	response, err := h.buildResponse(loc, fields)
	if err != nil {
		log.Printf("Error building response for IP %s: %v", ip, err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Error processing response")
		return
	}

	log.Printf("Successfully built response for IP: %s", ip)

	// Return the JSON response
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (h *IPHandler) buildResponse(loc *models.Location, fields string) (map[string]interface{}, error) {
	var response map[string]interface{}
	data, err := json.Marshal(loc)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	// Handle partial field selection
	if fields != "" {
		requestedFields := strings.Split(fields, ",")
		filteredResponse := make(map[string]interface{})
		for _, field := range requestedFields {
			field = strings.TrimSpace(field)
			if utils.Contains(h.config.AllowedFields, field) {
				filteredResponse[field] = response[field]
			}
		}
		response = filteredResponse
	}

	return response, nil
}
