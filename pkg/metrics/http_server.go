package metrics

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"network-sec-micro/pkg/health"
)

// StartMetricsServer starts an HTTP server for metrics and health checks
func StartMetricsServer(port string, healthHandler *health.Handler) error {
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Health check endpoints
	if healthHandler != nil {
		mux.HandleFunc("/health", healthHandler.Health)
		mux.HandleFunc("/ready", healthHandler.Ready)
		mux.HandleFunc("/live", healthHandler.Live)
	}

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Metrics server starting on port %s", port)
	return server.ListenAndServe()
}

// StartMetricsServerWithContext starts metrics server with context for graceful shutdown
func StartMetricsServerWithContext(ctx context.Context, port string, healthHandler *health.Handler) error {
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Health check endpoints
	if healthHandler != nil {
		mux.HandleFunc("/health", healthHandler.Health)
		mux.HandleFunc("/ready", healthHandler.Ready)
		mux.HandleFunc("/live", healthHandler.Live)
	}

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Printf("Metrics server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		log.Println("Shutting down metrics server...")
		return server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return fmt.Errorf("metrics server error: %w", err)
	}
}

