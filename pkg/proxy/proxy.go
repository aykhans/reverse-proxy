package proxy

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/natigmaderov/reverse-proxy/pkg/config"
	"github.com/natigmaderov/reverse-proxy/pkg/ratelimit"
	"github.com/natigmaderov/reverse-proxy/pkg/webhook"
)

// ReverseProxy wraps the standard reverse proxy with additional functionality
type ReverseProxy struct {
	proxy       *httputil.ReverseProxy
	config      *config.Config
	rateLimiter *ratelimit.RateLimiter
	notifier    *webhook.Notifier
}

// NewReverseProxy creates a new reverse proxy with custom settings
func NewReverseProxy(cfg *config.Config) (*ReverseProxy, error) {
	backend, err := url.Parse(cfg.BackendURL)
	if err != nil {
		return nil, err
	}

	// Create custom transport
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          cfg.MaxIdleConns,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   cfg.MaxIdleConns,
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(backend)
	proxy.Transport = transport

	// Customize the director
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Scheme = backend.Scheme
		req.URL.Host = backend.Host
		req.Host = backend.Host
	}

	// Create notifier
	notifier := webhook.NewNotifier(cfg.WebhookURL, cfg.ContainerID)

	// Create custom reverse proxy
	rp := &ReverseProxy{
		proxy:       proxy,
		config:      cfg,
		rateLimiter: ratelimit.NewRateLimiter(cfg.RateLimit),
		notifier:    notifier,
	}

	// Add error handler
	proxy.ErrorHandler = rp.errorHandler

	return rp, nil
}

// errorHandler handles proxy errors
func (rp *ReverseProxy) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	if err == context.Canceled {
		// Client canceled the request, log at debug level
		log.Printf("Client canceled request: %s", r.URL.Path)
		return
	}
	log.Printf("Proxy error: %v", err)
	http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
}

// ServeHTTP implements the http.Handler interface
func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get client IP
	clientIP := getClientIP(r)

	// Get limiter and tracker
	limiter, tracker := rp.rateLimiter.GetLimiter(clientIP)

	// Track request and get current rate
	currentRPS := tracker.TrackRequest()

	// Only check rate limit if we're actually exceeding the limit
	if currentRPS > rp.config.RateLimit {
		if !limiter.Allow() {
			log.Printf("Rate limit exceeded for %s - Current RPS: %d, Limit: %d",
				clientIP, currentRPS, rp.config.RateLimit)

			// Send webhook notification
			if err := rp.notifier.NotifyRateLimit(rp.config.RateLimit, currentRPS); err != nil {
				log.Printf("Failed to send webhook notification: %v", err)
			}

			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), rp.config.ProxyTimeout)
	defer cancel()

	// Serve with timeout context
	rp.proxy.ServeHTTP(w, r.WithContext(ctx))
}

// Close cleans up resources
func (rp *ReverseProxy) Close() {
	rp.rateLimiter.Close()
}

// getClientIP gets the real client IP considering X-Forwarded-For and X-Real-IP headers
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
