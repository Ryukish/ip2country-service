# `ip2country-service` Deployment Guide

This guide provides comprehensive instructions for deploying the `ip2country-service` in various modes, including local and distributed setups. It covers running the service using Docker, Docker Compose, and provides details about configuration environment variables, the rate limiting algorithm used, accessing Prometheus and Grafana for monitoring, trade-offs made during development, and potential future improvements.

---

## Table of Contents

- [Deployment Modes](#deployment-modes)
  - [Running in Local Mode via Dockerfile](#running-in-local-mode-via-dockerfile)
  - [Running with Docker Compose](#running-with-docker-compose)
  - [Local Mode](#local-mode)
  - [Distributed Mode](#distributed-mode)
  - [JSON or CSV Local Mode](#json-or-csv-local-mode)
- [Configuration Environment Variables](#configuration-environment-variables)
- [Rate Limiting Algorithm](#rate-limiting-algorithm)
- [Accessing Prometheus and Grafana Dashboards](#accessing-prometheus-and-grafana-dashboards)
  - [Prometheus Setup and Access](#prometheus-setup-and-access)
  - [Grafana Setup and Access](#grafana-setup-and-access)
  - [Grafana Dashboard Configuration](#grafana-dashboard-configuration)
- [Trade-offs](#trade-offs)
- [Future Improvements](#future-improvements)
- [Additional Code References](#additional-code-references)
- [Conclusion](#conclusion)
- [Additional Notes](#additional-notes)

---

## Deployment Modes

### Running in Local Mode via Dockerfile

You can run the `ip2country-service` in local mode using the provided Dockerfile. This method builds the application into a Docker image and runs it as a container without relying on Docker Compose. It's ideal for simple setups and quick testing.

#### Steps:

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/ryukish/ip2country-service.git
   cd ip2country-service
   ```

2. **Prepare the Data File**:

   Ensure you have the `ip_database.json` or `ip_database.csv` file in the `data` directory. If not, place your IP data file there.

3. **Review the Dockerfile**:

   The Dockerfile provided builds the application and includes the data files.

   **Dockerfile**:

   ```dockerfile
   # Build Stage
   FROM golang:1.23-alpine AS builder

   WORKDIR /app

   COPY go.mod go.sum ./
   RUN go mod download

   COPY . .

   RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ip2country-service ./cmd/server/main.go

   # Run Stage
   FROM alpine:latest

   WORKDIR /app

   COPY --from=builder /app/ip2country-service .
   COPY --from=builder /app/data/ ./data/

   EXPOSE 8080

   CMD ["./ip2country-service"]
   ```

   - **Build Stage**:
     - Uses the Go Alpine image to compile the application.
     - Downloads dependencies and builds the binary.
   - **Run Stage**:
     - Uses a minimal Alpine image.
     - Copies the binary and data files from the build stage.
     - Exposes port `8080` and runs the application.

4. **Build the Docker Image**:

   Run the following command to build the Docker image:

   ```bash
   docker build -t ip2country-service:latest .
   ```

5. **Run the Docker Container**:

   Run the container using the built image:

   ```bash
   docker run -d -p 8080:8080 \
     -e IP_DATABASE_TYPE=json \
     -e IP_DATABASE_PATH=/app/data/ip_database.json \
     -e RATE_LIMITER_TYPE=local \
     -e RATE_LIMIT=5 \
     -e RATE_CAPACITY=10 \
     -e RATE_JITTER=100 \
     -v ${PWD}/data:/app/data \
     ip2country-service:latest
   ```

   - **Environment Variables**:
     - `IP_DATABASE_TYPE=json`: Uses the JSON file as the database.
     - `IP_DATABASE_PATH=/app/data/ip_database.json`: Path inside the container to the data file.
     - `RATE_LIMITER_TYPE=local`: Uses the local rate limiter.
   - **Volume Mount**:
     - Mounts the local `data` directory to `/app/data` in the container.
   - **Port Mapping**:
     - Maps port `8080` of the container to port `8080` of the host.

6. **Verify the Service is Running**:

   Check that the container is running:

   ```bash
   docker ps
   ```

   You should see the `ip2country-service` container listed.

7. **Access the Service**:

   The service is now running and accessible at `http://localhost:8080`.

#### Testing the API

You can test the API using `curl` or any HTTP client:

```bash
curl 'http://localhost:8080/api/v1/find-country?ip=8.8.8.8'
```

---

### Running with Docker Compose

For more complex setups, especially when running multiple services like MongoDB, Redis, Prometheus, and Grafana, you can use Docker Compose. This method is suitable for both local development and production environments.

#### Steps:

1. **Ensure Docker and Docker Compose are Installed**:

   Verify installations:

   ```bash
   docker --version
   docker-compose --version
   ```

2. **Clone the Repository**:

   ```bash
   git clone https://github.com/ryukish/ip2country-service.git
   cd ip2country-service
   ```

3. **Review the `docker-compose.yml` File**:

   The provided `docker-compose.yml` includes services for MongoDB, Redis, the `ip2country-service`, data migration, Prometheus, and Grafana.

   **docker-compose.yml**:

   ```yaml
   version: '3.8'

   services:
     mongo:
       image: mongo:latest
       container_name: mongo
       ports:
         - "27017:27017"
       volumes:
         - ./data/mongo:/data/db
       networks:
         - ip2country-net

     redis:
       image: redis:latest
       container_name: redis
       ports:
         - "6379:6379"
       networks:
         - ip2country-net

     ip2country-service:
       image: ip2country-service:latest
       container_name: ip2country-service
       build: .
       ports:
         - "8080:8080"
       environment:
         - MONGODB_URI=mongodb://mongo:27017
         - MONGODB_NAME=ip2country
         - IP_DATABASE_TYPE=mongodb
         - RATE_LIMITER_TYPE=redis
         - REDIS_ADDR=redis:6379
         - RATE_LIMIT=1
         - RATE_CAPACITY=1
         - RATE_JITTER=10
       depends_on:
         - mongo
         - redis
       networks:
         - ip2country-net

     data-migration:
       image: golang:1.23-alpine
       container_name: data-migration
       working_dir: /data
       entrypoint: /bin/sh -c "go mod tidy && go run migrate.go -file=ip_database.json -type=json"
       environment:
         - MONGODB_URI=mongodb://mongo:27017
         - MONGODB_NAME=ip2country
       volumes:
         - ./data:/data
       networks:
         - ip2country-net
       depends_on:
         - mongo
       restart: "on-failure"

     prometheus:
       image: prom/prometheus
       container_name: prometheus
       volumes:
         - ./prometheus.yml:/etc/prometheus/prometheus.yml
       ports:
         - "9090:9090"
       networks:
         - ip2country-net

     grafana:
       image: grafana/grafana
       container_name: grafana
       ports:
         - "3000:3000"
       environment:
         - GF_SECURITY_ADMIN_PASSWORD=admin
       volumes:
         - ./grafana/provisioning/:/etc/grafana/provisioning/
         - ./grafana/dashboards/:/var/lib/grafana/dashboards/
       networks:
         - ip2country-net

   networks:
     ip2country-net:
   ```

4. **Build the Docker Image**:

   ```bash
   docker-compose build
   ```

5. **Run Docker Compose**:

   Start all services:

   ```bash
   docker-compose up
   ```

   To run in detached mode (in the background):

   ```bash
   docker-compose up -d
   ```

6. **Access the Services**:

   - **ip2country-service**: `http://localhost:8080`
   - **Prometheus**: `http://localhost:9090`
   - **Grafana**: `http://localhost:3000`

7. **Verify the Containers are Running**:

   ```bash
   docker-compose ps
   ```

   You should see all the services up and running.

---

### Local Mode

You can also run the service directly on your local machine without Docker, ideal for development and testing.

#### Steps:

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/ryukish/ip2country-service.git
   cd ip2country-service
   ```

2. **Install Dependencies**:

   Ensure you have Go installed and set up properly.

   ```bash
   go mod download
   ```

3. **Set Environment Variables** (optional):

   ```bash
   export IP_DATABASE_TYPE=json
   export IP_DATABASE_PATH=./data/ip_database.json
   export RATE_LIMITER_TYPE=local
   export RATE_LIMIT=5
   export RATE_CAPACITY=10
   export RATE_JITTER=100
   ```

4. **Run the Service**:

   ```bash
   go run cmd/server/main.go
   ```

5. **Access the Service**:

   The service will be available at `http://localhost:8080`.

---

### Distributed Mode

For production environments requiring scalability, run the service with MongoDB and Redis.

#### Steps:

1. **Set Up MongoDB and Redis**:

   Ensure you have MongoDB and Redis instances running.

2. **Update Configuration**:

   Set environment variables to point to your MongoDB and Redis instances.

   ```bash
   export IP_DATABASE_TYPE=mongodb
   export MONGODB_URI=mongodb://<username>:<password>@<host>:<port>
   export MONGODB_NAME=ip2country

   export RATE_LIMITER_TYPE=redis
   export REDIS_ADDR=<host>:<port>
   export REDIS_PASSWORD=<password>
   export REDIS_DB=0
   ```

3. **Run the Service**:

   ```bash
   go run cmd/server/main.go
   ```

4. **Access the Service**:

   The service will be available at `http://<your-server-ip>:8080`.

---

### JSON or CSV Local Mode

Run the service using JSON or CSV files as the data source without databases.

#### Steps:

1. **Prepare the Data File**:

   Place your `ip_database.json` or `ip_database.csv` file in the `data` directory.

2. **Update Configuration**:

   ```bash
   export IP_DATABASE_TYPE=json   # or 'csv'
   export IP_DATABASE_PATH=./data/ip_database.json   # or './data/ip_database.csv'
   ```

3. **Run the Service**:

   ```bash
   go run cmd/server/main.go
   ```

4. **Access the Service**:

   The service will be available at `http://localhost:8080`.

---

## Configuration Environment Variables

The `ip2country-service` uses several environment variables to configure its behavior. Here's a breakdown of each:

- **Database Configuration**:

  - `IP_DATABASE_TYPE`: Specifies the type of database to use. Options include `json`, `csv`, or `mongodb`.
  - `IP_DATABASE_PATH`: Path to the JSON or CSV file containing IP data. Used when `IP_DATABASE_TYPE` is `json` or `csv`.
  - `MONGODB_URI`: URI for connecting to the MongoDB instance (used when `IP_DATABASE_TYPE` is `mongodb`).
  - `MONGODB_NAME`: Name of the MongoDB database to use.

- **Rate Limiter Configuration**:

  - `RATE_LIMITER_TYPE`: Determines the rate limiting strategy. Options include `local` or `redis`.
  - `RATE_LIMIT`: The maximum number of requests allowed per time window.
  - `RATE_CAPACITY`: The capacity of the rate limiter bucket.
  - `RATE_JITTER`: Adds randomness to the rate limiting to prevent bursts of requests.
  - `REDIS_ADDR`: Address of the Redis server (used when `RATE_LIMITER_TYPE` is `redis`).
  - `REDIS_PASSWORD`: Password for the Redis server, if required.
  - `REDIS_DB`: Redis database number to use.

- **Service Configuration**:

  - `PORT`: The port on which the service will listen (default is `8080`).
  - `ALLOWED_FIELDS`: Comma-separated list of fields that can be returned in the API response.

---

## Rate Limiting Algorithm

The `ip2country-service` employs a **token bucket algorithm** for rate limiting. This algorithm efficiently controls the rate at which requests are processed, ensuring fair usage and preventing abuse.

### How It Works:

- **Token Bucket**: A bucket is filled with tokens at a steady rate. Each incoming request consumes one token. If the bucket has no tokens, the request is denied (rate-limited).

- **Capacity**: The bucket has a maximum capacity (`RATE_CAPACITY`), which limits the maximum burst of requests that can be handled instantly.

- **Refill Rate**: Tokens are added to the bucket at a rate defined by `RATE_LIMIT`. For example, if `RATE_LIMIT` is 5, five tokens are added per second.

- **Jitter**: `RATE_JITTER` introduces randomness to the refill interval, which helps to prevent synchronized bursts of traffic, smoothing out the load on the service.

### Local vs. Redis Rate Limiter:

- **Local Rate Limiter**: Suitable for single-instance deployments. The rate limiting is enforced per instance and does not synchronize across multiple instances.

- **Redis Rate Limiter**: Ideal for distributed environments where multiple instances of the service are running. Redis acts as a centralized store to synchronize the token buckets across all instances.

---

## Accessing Prometheus and Grafana Dashboards

To monitor the `ip2country-service`, you can use Prometheus for metrics collection and Grafana for visualization.

### Prometheus Setup and Access

#### Steps:

1. **Ensure Prometheus is Running**:

   Verify that the Prometheus container is up and running:

   ```bash
   docker-compose ps
   ```

   You should see a service named `prometheus` running.

2. **Access Prometheus Web UI**:

   Open your web browser and navigate to:

   ```
   http://localhost:9090
   ```

   This will bring up the Prometheus web interface where you can query metrics and check the status of your targets.

3. **Prometheus Configuration (`prometheus.yml`)**:

   `prometheus.yml` file is set up correctly. It should be mounted in the Prometheus container.

   **prometheus.yml**:

   ```yaml
   global:
     scrape_interval: 2s

   scrape_configs:
     - job_name: 'ip2country-service'
       static_configs:
         - targets: ['ip2country-service:8080']
   ```

   - **Note**: Since all services are on the same Docker network (`ip2country-net`), Prometheus can access `ip2country-service` by its service name.

4. **Verify Targets in Prometheus**:

   Go to the **Targets** page in Prometheus:

   ```
   http://localhost:9090/targets
   ```

   Ensure that the `ip2country-service` target is **UP**.

### Grafana Setup and Access

#### Steps:

1. **Ensure Grafana is Running**:

   Verify that the Grafana container is up and running:

   ```bash
   docker-compose ps
   ```

   You should see a service named `grafana` running.

2. **Access Grafana Web UI**:

   Open your web browser and navigate to:

   ```
   http://localhost:3000
   ```

3. **Log in to Grafana**:

   - **Username**: `admin`
   - **Password**: `admin` (as set in `GF_SECURITY_ADMIN_PASSWORD`)

   You will be prompted to change the password upon first login.

4. **Verify Prometheus Data Source**:

   The Prometheus data source should be automatically configured via provisioning. You can verify this by navigating to:

   - **Configuration** (gear icon) -> **Data Sources**

   You should see a data source named **Prometheus** pointing to `http://prometheus:9090`.

### Grafana Dashboard Configuration

To visualize metrics, you'll need to set up a Grafana dashboard.

#### Steps:

1. **Create Directory Structure for Provisioning**:

   ```bash
   mkdir -p grafana/provisioning/datasources
   mkdir -p grafana/provisioning/dashboards
   mkdir -p grafana/dashboards
   ```

2. **Configure Data Source Provisioning**:

   **File**: `grafana/provisioning/datasources/datasource.yml`

   ```yaml
   apiVersion: 1

   datasources:
     - name: Prometheus
       type: prometheus
       access: proxy
       url: http://prometheus:9090
       isDefault: true
   ```

3. **Configure Dashboard Provisioning**:

   **File**: `grafana/provisioning/dashboards/dashboard.yml`

   ```yaml
   apiVersion: 1

   providers:
     - name: 'default'
       orgId: 1
       folder: ''
       type: file
       options:
         path: /var/lib/grafana/dashboards
         disableDeletion: false
         updateIntervalSeconds: 10
   ```

4. **Create Grafana Dashboard JSON**:

   **File**: `grafana/dashboards/ip2country_dashboard.json`

   ```json
   {
     "title": "ip2country-service Metrics",
     "panels": [
       {
         "type": "graph",
         "title": "Total HTTP Requests",
         "targets": [
           {
             "expr": "sum(rate(http_requests_total[1m])) by (path)",
             "legendFormat": "{{path}}",
             "refId": "A"
           }
         ],
         "id": 1
       },
       {
         "type": "graph",
         "title": "Request Duration (95th Percentile)",
         "targets": [
           {
             "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path))",
             "legendFormat": "{{path}}",
             "refId": "B"
           }
         ],
         "id": 2
       },
       {
         "type": "graph",
         "title": "Rate Limit Exceeded",
         "targets": [
           {
             "expr": "sum(rate(http_rate_limit_exceeded_total[1m])) by (path)",
             "legendFormat": "{{path}}",
             "refId": "C"
           }
         ],
         "id": 3
       }
     ],
     "schemaVersion": 16,
     "version": 0
   }
   ```

5. **Update `docker-compose.yml` for Grafana**:

   Ensure your Grafana service is set up to load the dashboard and configure the Prometheus data source.

   ```yaml
   grafana:
     image: grafana/grafana
     container_name: grafana
     ports:
       - "3000:3000"
     environment:
       - GF_SECURITY_ADMIN_PASSWORD=admin
     volumes:
       - ./grafana/provisioning/:/etc/grafana/provisioning/
       - ./grafana/dashboards/:/var/lib/grafana/dashboards/
     networks:
       - ip2country-net
   ```

6. **Access the Grafana Dashboard**:

   - Navigate to **Dashboards** (four squares icon) -> **Manage**.
   - Open the `ip2country-service Metrics` dashboard.

7. **Generate Traffic to See Metrics**:

   Send requests to your `ip2country-service` to generate metrics:

   ```bash
   for i in {1..50}; do
     curl 'http://localhost:8080/api/v1/find-country?ip=8.8.8.8'
   done
   ```

8. **Refresh Grafana Dashboard**:

   - Adjust the time range to the last 5 or 15 minutes.
   - Set the refresh interval (e.g., every 10 seconds).

9. **Verify Metrics are Displayed**:

   You should now see graphs showing the total HTTP requests, request duration, and rate limit exceeded counts.

---

## Trade-offs

- **Database Choice**:

  - *MongoDB* provides scalability and flexibility but requires more resources and setup.
  - *JSON/CSV Files* are easy to manage and deploy but are not suitable for large-scale or high-concurrency environments.

- **Rate Limiting Strategy**:

  - *Local Rate Limiter* is simple and efficient for single-instance deployments but doesn't prevent abuse across multiple instances.
  - *Redis Rate Limiter* offers distributed rate limiting but introduces additional network latency and complexity.

- **Deployment Complexity**:

  - *Docker vs. Docker Compose*:
    - Docker is simpler for single-service deployments.
    - Docker Compose is better for multi-service setups but adds complexity.
  - *Manual Setup* offers more control but requires more effort and is prone to configuration errors.

---

## Future Improvements

- **Enhanced Security**:

  - Implement authentication mechanisms for the API.
  - Enable TLS/SSL encryption for data in transit.
  - Secure MongoDB and Redis instances with proper authentication and network policies.

- **Performance Optimization**:

  - Implement caching strategies (e.g., in-memory caching) to reduce database load.
  - Optimize database queries and indexing.

- **Scalability**:

  - Integrate with container orchestration platforms like Kubernetes for better scalability and high availability.
  - Use load balancers to distribute traffic across multiple service instances.

- **Monitoring and Logging**:

  - Integrate with centralized logging systems (e.g., ELK stack) for better log management.
  - Enhance monitoring with tools like Grafana to visualize Prometheus metrics.

- **API Enhancements**:

  - Add support for IPv6 addresses.
  - Implement pagination and filtering for APIs that return lists.
  - Provide more detailed error messages and status codes.

---

## Additional Code References

### Database Initialization

The service supports multiple types of databases. The appropriate database handler is selected based on the `IP_DATABASE_TYPE` configuration.

```go
package database

import (
  "fmt"
  "ip2country-service/config"
  "ip2country-service/internal/models"
)

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
```

### Data Models

The `IPLocation` struct represents the data model for an IP location entry.

```go
type IPLocation struct {
  IPFrom  uint32 `json:"ip_from"`
  IPTo    uint32 `json:"ip_to"`
  Country string `json:"country"`
  Region  string `json:"region"`
  City    string `json:"city"`
}
```

### Monitoring Package Metrics

The monitoring package defines Prometheus metrics.

**File**: `monitoring/monitoring.go`

```go
package monitoring

import (
  "github.com/prometheus/client_golang/prometheus"
)

var (
  RequestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
      Name: "http_requests_total",
      Help: "Total number of HTTP requests",
    },
    []string{"path"},
  )

  RequestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
      Name:    "http_request_duration_seconds",
      Help:    "Duration of HTTP requests in seconds",
      Buckets: prometheus.DefBuckets,
    },
    []string{"path"},
  )

  RateLimitExceeded = prometheus.NewCounterVec(
    prometheus.CounterOpts{
      Name: "http_rate_limit_exceeded_total",
      Help: "Total number of HTTP requests that exceeded the rate limit",
    },
    []string{"path"},
  )
)

func init() {
  prometheus.MustRegister(RequestsTotal, RequestDuration, RateLimitExceeded)
}
```

---

## Conclusion

Feel free to customize the configurations and reach out if you need further assistance.
---

## Additional Notes

- **Data Migration Service**:

  The `data-migration` service in the `docker-compose.yml` is responsible for importing IP data into MongoDB.

  - **Entry Point**:

    ```yaml
    entrypoint: /bin/sh -c "go mod tidy && go run migrate.go -file=ip_database.json -type=json"
    ```

    - This runs a Go script `migrate.go` to import data.

- **Monitoring Tools**:

  - **Prometheus** and **Grafana** are included for monitoring.

  - **Prometheus Configuration** (`prometheus.yml`):

    Ensure you have a `prometheus.yml` file to configure Prometheus to scrape metrics from your services.

  - **Grafana Dashboard**:

    The `grafana_dashboard.json` file is used to set up a pre-configured dashboard in Grafana.

- **Networking**:

  All services are connected via the `ip2country-net` network for internal communication.

- **Testing and Troubleshooting**:

  - If you encounter issues with the Grafana dashboard not displaying data:
    - Ensure you're generating traffic to the `ip2country-service`.
    - Verify Prometheus is scraping metrics correctly.
    - Check the metrics endpoint (`http://localhost:8080/metrics`) for exposed metrics.
    - Confirm that the Grafana queries match the metric names and labels.
