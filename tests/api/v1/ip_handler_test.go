package v1_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "ip2country-service/api/v1"
	"ip2country-service/config"
	"ip2country-service/internal/models"
)

type mockDatabase struct{}

func (m *mockDatabase) Find(ip string) (*models.Location, error) {
	if ip == "10.0.0.1" {
		return &models.Location{
			Country: "US",
			Region:  "California",
			City:    "Los Angeles",
		}, nil
	}
	return nil, fmt.Errorf("IP not found")
}

func TestGetLocation(t *testing.T) {
	db := &mockDatabase{}
	cfg := &config.Config{
		AllowedFields: []string{"country", "region", "city"},
	}
	handler := v1.NewIPHandler(db, cfg)

	req, err := http.NewRequest("GET", "/find-country?ip=10.0.0.1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.GetLocation(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := map[string]string{
		"country": "US",
		"region":  "California",
		"city":    "Los Angeles",
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("could not parse response: %v", err)
	}

	if !equal(response, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", response, expected)
	}
}

func equal(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
