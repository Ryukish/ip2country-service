package database_test

import (
	"ip2country-service/internal/database"
	"ip2country-service/internal/models"
	"testing"
)

func TestCSVDatabase_Find(t *testing.T) {
	// Mock data for testing
	mockData := []database.IPLocation{
		{IPFrom: 167772160, IPTo: 167772175, Country: "US", Region: "California", City: "Los Angeles"},
		{IPFrom: 167772176, IPTo: 167772191, Country: "US", Region: "New York", City: "New York"},
	}

	db := &database.CSVDatabase{
		DatabaseLocal: database.DatabaseLocal{Locations: mockData},
	}

	tests := []struct {
		ip      string
		want    *models.Location
		wantErr bool
	}{
		{"10.0.0.1", &models.Location{Country: "US", Region: "California", City: "Los Angeles"}, false},
		{"10.0.0.16", &models.Location{Country: "US", Region: "New York", City: "New York"}, false},
		{"10.0.0.32", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got, err := db.Find(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (got.Country != tt.want.Country || got.Region != tt.want.Region || got.City != tt.want.City) {
				t.Errorf("Find() = %v, want %v", got, tt.want)
			}
		})
	}
}
