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
