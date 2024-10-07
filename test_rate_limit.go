package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	// Configuration
	url := "http://localhost:8080/api/v1/locations?ip=10.0.0.1"
	totalRequests := 100 // Total requests to send

	// Function to send requests
	for i := 1; i <= totalRequests; i++ {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Request %d: Error %v\n", i, err)
			continue
		}
		fmt.Printf("Request %d: HTTP Status %s\n", i, resp.Status)
		resp.Body.Close()
		time.Sleep(100 * time.Millisecond) // Send requests every 100ms

	}
}
