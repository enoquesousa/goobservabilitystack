// Command server is the entrypoint for the goobservabilitystack service.
// It starts an application HTTP server on HTTP_PORT (default 8080) and a
// dedicated metrics HTTP server on METRICS_PORT (default 9091).
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ESousa97/goobservabilitystack/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// envOr returns the value of the environment variable named by key,
// or fallback if the variable is unset or empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	appPort := envOr("HTTP_PORT", "8080")
	metricsPort := envOr("METRICS_PORT", "9091")

	reg, err := metrics.NewRegistry()
	if err != nil {
		log.Fatalf("failed to initialize metrics registry: %v", err)
	}

	worker, err := metrics.NewWorker(reg.EventsCounter(), 5*time.Second)
	if err != nil {
		log.Fatalf("failed to initialize metrics worker: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	go worker.Run(ctx)

	// Metrics server — exposes /metrics for Prometheus scraping.
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.HandlerFor(reg.Gatherer(), promhttp.HandlerOpts{}))
	metricsServer := &http.Server{
		Addr:         ":" + metricsPort,
		Handler:      metricsMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Application server — serves application routes and health checks.
	appMux := http.NewServeMux()
	appMux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	appServer := &http.Server{
		Addr:         ":" + appPort,
		Handler:      appMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("metrics server listening on :%s", metricsPort)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("metrics server error: %v", err)
		}
	}()

	go func() {
		log.Printf("app server listening on :%s", appPort)
		if err := appServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("app server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received, draining connections...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("metrics server shutdown error: %v", err)
	}
	if err := appServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("app server shutdown error: %v", err)
	}

	log.Println("shutdown complete")
}
