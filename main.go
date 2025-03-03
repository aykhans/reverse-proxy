package main

import (
	"log"
	"net/http"

	"github.com/natigmaderov/reverse-proxy/pkg/config"
	"github.com/natigmaderov/reverse-proxy/pkg/proxy"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting reverse proxy → Backend: %s | Rate Limit: %d requests/sec | Container ID: %s",
		cfg.BackendURL, cfg.RateLimit, cfg.ContainerID)

	// Create custom reverse proxy
	reverseProxy, err := proxy.NewReverseProxy(cfg)
	if err != nil {
		log.Fatalf("Failed to create reverse proxy: %v", err)
	}
	defer reverseProxy.Close()

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":8080",
		Handler:      reverseProxy,
		ReadTimeout:  cfg.ProxyTimeout,
		WriteTimeout: cfg.ProxyTimeout,
		IdleTimeout:  cfg.IdleConnTimeout,
	}

	log.Println("Reverse proxy with rate limiting running on :8080")
	log.Fatal(server.ListenAndServe())
}
