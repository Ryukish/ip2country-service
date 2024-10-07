package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port            string
	RateLimit       int
	DatabaseType    string // "json" or "mongodb"
	DatabasePath    string // For JSON files
	MongoDBURI      string // For MongoDB connection
	MongoDBName     string
	RateLimiterType string // "local" or "redis"
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	AllowedFields   []string // Fields allowed for partial retrieval
}

// LoadConfig loads the configuration from environment variables or defaults
func LoadConfig() *Config {
	rateLimit, err := strconv.Atoi(os.Getenv("RATE_LIMIT"))
	if err != nil {
		log.Println("Invalid RATE_LIMIT, defaulting to 5")
		rateLimit = 5
	}

	return &Config{
		Port:            getEnv("PORT", "8080"),
		RateLimit:       rateLimit,
		DatabaseType:    getEnv("IP_DATABASE_TYPE", "json"),
		DatabasePath:    getEnv("IP_DATABASE_PATH", "/data/ip_database.json"),
		MongoDBURI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBName:     getEnv("MONGODB_NAME", "ip2country"),
		RateLimiterType: getEnv("RATE_LIMITER_TYPE", "local"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         getEnvAsInt("REDIS_DB", 0),
		AllowedFields:   []string{"country", "city"}, // Default allowed fields for partial retrieval
	}
}

// getEnv retrieves the value of the environment variable named by the key.
// It returns the value or the defaultValue if the variable is not present.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves the value of the environment variable as an integer.
// It returns the value or the defaultValue if the variable is not present or invalid.
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
