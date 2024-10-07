package config_test

import (
	"os"
	"testing"

	"ip2country-service/config"
)

func TestLoadConfig(t *testing.T) {
	// Set up environment variables for testing
	os.Setenv("PORT", "9090")
	os.Setenv("RATE_LIMIT", "10")
	os.Setenv("IP_DATABASE_TYPE", "mongodb")
	os.Setenv("IP_DATABASE_PATH", "/custom/path/to/db.json")
	os.Setenv("MONGODB_URI", "mongodb://customhost:27017")
	os.Setenv("MONGODB_NAME", "customdb")
	os.Setenv("RATE_LIMITER_TYPE", "redis")
	os.Setenv("REDIS_ADDR", "customhost:6379")
	os.Setenv("REDIS_PASSWORD", "custompassword")
	os.Setenv("REDIS_DB", "1")

	// Load the configuration
	config := config.LoadConfig()

	// Test the loaded configuration
	if config.Port != "9090" {
		t.Errorf("Expected Port to be '9090', got '%s'", config.Port)
	}
	if config.RateLimit != 10 {
		t.Errorf("Expected RateLimit to be 10, got %d", config.RateLimit)
	}
	if config.DatabaseType != "mongodb" {
		t.Errorf("Expected DatabaseType to be 'mongodb', got '%s'", config.DatabaseType)
	}
	if config.DatabasePath != "/custom/path/to/db.json" {
		t.Errorf("Expected DatabasePath to be '/custom/path/to/db.json', got '%s'", config.DatabasePath)
	}
	if config.MongoDBURI != "mongodb://customhost:27017" {
		t.Errorf("Expected MongoDBURI to be 'mongodb://customhost:27017', got '%s'", config.MongoDBURI)
	}
	if config.MongoDBName != "customdb" {
		t.Errorf("Expected MongoDBName to be 'customdb', got '%s'", config.MongoDBName)
	}
	if config.RateLimiterType != "redis" {
		t.Errorf("Expected RateLimiterType to be 'redis', got '%s'", config.RateLimiterType)
	}
	if config.RedisAddr != "customhost:6379" {
		t.Errorf("Expected RedisAddr to be 'customhost:6379', got '%s'", config.RedisAddr)
	}
	if config.RedisPassword != "custompassword" {
		t.Errorf("Expected RedisPassword to be 'custompassword', got '%s'", config.RedisPassword)
	}
	if config.RedisDB != 1 {
		t.Errorf("Expected RedisDB to be 1, got %d", config.RedisDB)
	}
	if len(config.AllowedFields) != 2 || config.AllowedFields[0] != "country" || config.AllowedFields[1] != "city" {
		t.Errorf("Expected AllowedFields to be ['country', 'city'], got %v", config.AllowedFields)
	}

	// Clean up environment variables
	os.Unsetenv("PORT")
	os.Unsetenv("RATE_LIMIT")
	os.Unsetenv("IP_DATABASE_TYPE")
	os.Unsetenv("IP_DATABASE_PATH")
	os.Unsetenv("MONGODB_URI")
	os.Unsetenv("MONGODB_NAME")
	os.Unsetenv("RATE_LIMITER_TYPE")
	os.Unsetenv("REDIS_ADDR")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("REDIS_DB")
}
