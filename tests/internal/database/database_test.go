package database_test

import (
	"ip2country-service/config"
	"ip2country-service/internal/database"
	"testing"
)

func TestNewIPDatabase(t *testing.T) {
	csvFilePath := "ip_database.csv"
	jsonFilePath := "ip_database.json"

	t.Run("CSV_Database", func(t *testing.T) {
		config := &config.Config{
			DatabaseType: "csv",
			DatabasePath: csvFilePath,
		}
		_, err := database.NewIPDatabase(config)
		if err != nil {
			t.Errorf("NewIPDatabase() error = %v, expectedError false", err)
		}
	})

	t.Run("JSON_Database", func(t *testing.T) {
		config := &config.Config{
			DatabaseType: "json",
			DatabasePath: jsonFilePath,
		}
		_, err := database.NewIPDatabase(config)
		if err != nil {
			t.Errorf("NewIPDatabase() error = %v, expectedError false", err)
		}
	})
}
