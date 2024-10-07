#!/bin/bash

# Configuration
URL="http://localhost:8080/api/v1/locations?ip=10.0.0.1"
RATE_LIMIT=5  # Number of allowed requests
TOTAL_REQUESTS=10  # Total requests to send

# Function to send requests
send_requests() {
  for ((i=1; i<=TOTAL_REQUESTS; i++)); do
    response=$(curl -s -o /dev/null -w "%{http_code}" $URL)
    echo "Request $i: HTTP Status $response"
    sleep 1  # Adjust sleep time if needed
  done
}

# Run the test
send_requests