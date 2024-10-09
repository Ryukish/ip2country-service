package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"ip2country-service/config"
	"ip2country-service/internal/database"
	"ip2country-service/internal/models"
	"ip2country-service/monitoring"
	"ip2country-service/pkg/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

type IPHandler struct {
	db     database.IPDatabase
	config *config.Config
	cache  *cache.Cache
}

func NewIPHandler(db database.IPDatabase, cfg *config.Config) *IPHandler {
	// Create a cache with a default expiration time of 5 minutes and purge unused items every 10 minutes
	c := cache.New(5*time.Minute, 10*time.Minute)
	return &IPHandler{db: db, config: cfg, cache: c}
}

func (h *IPHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now() // Start timing the request

	ip := r.URL.Query().Get("ip")
	fields := r.URL.Query().Get("fields")

	log.Printf("Received request for IP: %s with fields: %s", ip, fields)

	// Validate the IP
	if !utils.ValidateIP(ip) {
		log.Printf("Invalid IP address: %s", ip)
		monitoring.RequestsTotal.WithLabelValues(r.URL.Path, "error").Inc() // Increment request count
		monitoring.RateLimitExceeded.WithLabelValues(r.URL.Path).Inc()      // Increment rate limit exceeded count
		utils.RespondWithError(w, http.StatusBadRequest, utils.ErrInvalidIP.Error())
		return
	}

	var loc *models.Location
	var err error
	var cacheHit bool

	// Measure IP lookup time, including cache check
	ipLookupStart := time.Now()

	// Check cache first
	if cachedLoc, found := h.cache.Get(ip); found {
		log.Printf("IP found in cache: %s", ip)
		loc = cachedLoc.(*models.Location)
		cacheHit = true
	} else {
		// Query the database for the IP location
		log.Printf("Querying database for IP: %s", ip)
		loc, err = h.db.Find(ip)
		if err != nil {
			monitoring.RequestsTotal.WithLabelValues(r.URL.Path, "error").Inc()
			if errors.Is(err, utils.ErrIpNotFound) {
				log.Printf("IP not found in the database: %s", ip)
				utils.RespondWithError(w, http.StatusNotFound, err.Error())
			} else {
				log.Printf("Error querying database for IP %s: %v", ip, err)
				utils.RespondWithError(w, http.StatusInternalServerError, utils.ErrDatabaseQuery.Error())
			}
			return
		}
		// Cache the result
		h.cache.Set(ip, loc, cache.DefaultExpiration)
	}

	// Record IP lookup duration
	ipLookupDuration := time.Since(ipLookupStart)
	monitoring.IPLookupDuration.WithLabelValues().Observe(ipLookupDuration.Seconds())

	// Record cache hit/miss
	if cacheHit {
		monitoring.CacheHits.WithLabelValues(r.URL.Path).Inc()
	} else {
		monitoring.CacheMisses.WithLabelValues(r.URL.Path).Inc()
	}

	log.Printf("IP found: %+v", loc)

	// Build the response
	log.Printf("Building response for IP: %s", ip)
	response, err := h.buildResponse(loc, fields)
	if err != nil {
		monitoring.RequestsTotal.WithLabelValues(r.URL.Path, "error").Inc()
		log.Printf("Error building response for IP %s: %v", ip, err)
		if errors.Is(err, utils.ErrInvalidFields) {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, utils.ErrInternalServer.Error())
		}
		return
	}

	log.Printf("Successfully built response for IP: %s", ip)

	// Record the request duration
	duration := time.Since(startTime).Seconds()
	monitoring.RequestDuration.WithLabelValues(r.URL.Path).Observe(duration)

	// Increment the request count
	monitoring.RequestsTotal.WithLabelValues(r.URL.Path, "success").Inc()

	// Return the JSON response
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (h *IPHandler) buildResponse(loc *models.Location, fields string) (map[string]interface{}, error) {
	var response map[string]interface{}
	data, err := json.Marshal(loc)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", utils.ErrJSONMarshal, err)
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("%w: %v", utils.ErrJSONUnmarshal, err)
	}

	// Handle partial field selection
	if fields != "" {
		requestedFields := strings.Split(fields, ",")
		filteredResponse := make(map[string]interface{})
		for _, field := range requestedFields {
			field = strings.TrimSpace(field)
			if utils.Contains(h.config.AllowedFields, field) {
				filteredResponse[field] = response[field]
				monitoring.AllowedFieldsUsage.WithLabelValues(field).Inc()
			} else {
				return nil, fmt.Errorf("%w: %s", utils.ErrInvalidFields, field)
			}
		}
		response = filteredResponse
	}

	return response, nil
}
