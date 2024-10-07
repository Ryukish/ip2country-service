package models

type Location struct {
	Country string `json:"country,omitempty"`
	Region  string `json:"region,omitempty"`
	City    string `json:"city,omitempty"`
}
