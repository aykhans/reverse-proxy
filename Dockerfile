# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /reverse-proxy

# Final stage
FROM alpine:3.19

WORKDIR /

# Copy the binary from builder
COPY --from=builder /reverse-proxy /reverse-proxy

# Create non-root user
RUN adduser -D -H -h /app appuser
USER appuser

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/reverse-proxy"] 