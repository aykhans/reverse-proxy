# Rate-Limited Reverse Proxy

A highly configurable reverse proxy with per-client rate limiting and webhook notifications.

## Features

- Per-client IP rate limiting
- Configurable rate limits and timeouts
- Webhook notifications for rate limit violations
- Support for X-Forwarded-For and X-Real-IP headers
- Environment-based configuration
- Clean shutdown handling
- Efficient memory usage with automatic cleanup

## Installation

```bash
go get github.com/natigmaderov/reverse-proxy
```

## Configuration

The proxy can be configured using environment variables or a `.env` file. To get started:

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Modify the `.env` file with your settings:
```env
# Backend URL configuration
BACKEND_URL=http://localhost:8080
# Rate limiting configuration
RATE_LIMIT=10
# Container identification
CONTAINER_ID=local-dev
# Webhook configuration
WEBHOOK_URL=http://webhook.example.com/notify
# Proxy configuration
PROXY_TIMEOUT_SECONDS=30
MAX_IDLE_CONNS=100
IDLE_CONN_TIMEOUT_SECONDS=90
```

Note: The `.env` file is not tracked in git for security reasons. Make sure to keep your environment files secure and never commit them to version control.

## Usage

### As a standalone proxy

```go
package main

import (
    "log"
    "net/http"

    "github.com/natigmaderov/reverse-proxy/pkg/config"
    "github.com/natigmaderov/reverse-proxy/pkg/proxy"
)

func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    proxy, err := proxy.NewReverseProxy(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer proxy.Close()

    server := &http.Server{
        Addr:    ":8080",
        Handler: proxy,
    }

    log.Fatal(server.ListenAndServe())
}
```

### As a library

```go
import "github.com/natigmaderov/reverse-proxy/pkg/proxy"

// Create a new proxy with custom configuration
proxy, err := proxy.NewReverseProxy(&config.Config{
    BackendURL:      "http://backend-service:8080",
    RateLimit:       100,
    ContainerID:     "my-service",
    WebhookURL:      "http://monitoring:8080/webhook",
    ProxyTimeout:    30 * time.Second,
    MaxIdleConns:    100,
    IdleConnTimeout: 90 * time.Second,
})

// Use it as an http.Handler
http.Handle("/", proxy)
```

## Rate Limiting

The rate limiter uses a combination of:
- Token bucket algorithm for precise rate limiting
- Sliding window counter for accurate RPS tracking
- Per-client IP tracking
- Automatic cleanup of inactive clients

## Webhook Notifications

When rate limits are exceeded, the proxy sends a webhook notification with the following payload:

```json
{
    "container_id": "my-service",
    "limit_expected_rps": 10,
    "limit_exceeded_rps": 15,
    "message": "Rate limit exceeded"
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 