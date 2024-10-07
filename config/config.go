package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	RateLimit       float64
	DatabaseType    string // "json" or "mongodb"
	DatabasePath    string // For JSON files
	MongoDBURI      string // For MongoDB connection
	MongoDBName     string
	RateLimiterType string // "local" or "redis"
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	AllowedFields   []string // Fields allowed for partial retrieval
	RateCapacity    float64
	RateJitter      time.Duration
}

// LoadConfig loads the configuration from environment variables or defaults
func LoadConfig() *Config {

	return &Config{
		Port:            getEnv("PORT", "8080"),
		RateLimit:       getEnvAsFloat("RATE_LIMIT", 1),
		DatabaseType:    getEnv("IP_DATABASE_TYPE", "json"),
		DatabasePath:    getEnv("IP_DATABASE_PATH", "./data/ip_database.json"),
		MongoDBURI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBName:     getEnv("MONGODB_NAME", "ip2country"),
		RateLimiterType: getEnv("RATE_LIMITER_TYPE", "local"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         getEnvAsInt("REDIS_DB", 0),
		AllowedFields:   []string{"country", "city"}, // Default allowed fields for partial retrieval
		RateCapacity:    getEnvAsFloat("RATE_CAPACITY", 5),
		RateJitter:      time.Duration(getEnvAsInt("RATE_JITTER", 100)) * time.Millisecond,
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

// getEnvAsFloat retrieves the value of the environment variable as a float64.
// It returns the value or the defaultValue if the variable is not present or invalid.
func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}
