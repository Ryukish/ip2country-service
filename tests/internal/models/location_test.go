package models

import (
	"encoding/json"
	"ip2country-service/internal/models"
	"testing"
)

func TestLocationJSONMarshalling(t *testing.T) {
	location := models.Location{
		Country: "USA",
		Region:  "California",
		City:    "San Francisco",
	}

	data, err := json.Marshal(location)
	if err != nil {
		t.Fatalf("Failed to marshal Location: %v", err)
	}

	expectedJSON := `{"country":"USA","region":"California","city":"San Francisco"}`
	if string(data) != expectedJSON {
		t.Errorf("Expected JSON %s, but got %s", expectedJSON, string(data))
	}
}

func TestLocationJSONUnmarshalling(t *testing.T) {
	jsonData := `{"country":"USA","region":"California","city":"San Francisco"}`
	var location models.Location

	err := json.Unmarshal([]byte(jsonData), &location)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if location.Country != "USA" || location.Region != "California" || location.City != "San Francisco" {
		t.Errorf("Unmarshalled Location does not match expected values: %+v", location)
	}
}
