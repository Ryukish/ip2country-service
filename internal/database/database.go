package database

import (
	"fmt"
	"ip2country-service/config"
	"ip2country-service/internal/models"
)

type IPLocation struct {
	IPFrom  uint32 `json:"ip_from"`
	IPTo    uint32 `json:"ip_to"`
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
}

type DatabaseLocal struct {
	Locations []IPLocation
}

type IPDatabase interface {
	Find(ip string) (*models.Location, error)
}

func NewIPDatabase(cfg *config.Config) (IPDatabase, error) {
	switch cfg.DatabaseType {
	case "csv":
		return NewCSVDatabase(cfg.DatabasePath)
	case "json":
		return NewJSONDatabase(cfg.DatabasePath)
	case "mongodb":
		return NewMongoDatabase(cfg.MongoDBURI, cfg.MongoDBName)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DatabaseType)
	}
}
