# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
COPY shared/ ./shared/

# Download dependencies
RUN go mod download

# Copy source code
COPY services/auth-service/ ./services/auth-service/

# Build the application
WORKDIR /app/services/auth-service
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/auth-service ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/services/auth-service/bin/auth-service .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./auth-service"]