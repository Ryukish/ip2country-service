package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ip2country-service/api"
	"ip2country-service/config"
	"ip2country-service/internal/models"

	"github.com/gorilla/mux"
)

type mockDatabase struct{}

func (m *mockDatabase) Find(ip string) (*models.Location, error) {
	return nil, nil
}

func TestRegisterHandlers(t *testing.T) {
	router := mux.NewRouter()
	db := &mockDatabase{}
	cfg := &config.Config{}

	api.RegisterHandlers(router, db, cfg)

	tests := []struct {
		route  string
		method string
		want   int
	}{
		{"/locations?ip=invalid_ip", http.MethodGet, http.StatusBadRequest}, // Expecting 400 for invalid IP
		{"/metrics", http.MethodGet, http.StatusOK},
	}

	for _, tt := range tests {
		req, err := http.NewRequest(tt.method, tt.route, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != tt.want {
			t.Errorf("handler returned wrong status code: got %v want %v", status, tt.want)
		}
	}
}
