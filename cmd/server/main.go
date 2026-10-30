// Command server is the entrypoint for the goobservabilitystack service.
// It starts an application HTTP server on HTTP_HOST:HTTP_PORT (default 127.0.0.1:8080)
// and a metrics HTTP server on METRICS_HOST:METRICS_PORT (default 127.0.0.1:9091).
// OpenTelemetry tracing is initialised at startup and flushed on graceful shutdown.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/enoquesousa/goobservabilitystack/internal/config"
	"github.com/enoquesousa/goobservabilitystack/internal/metrics"
	"github.com/enoquesousa/goobservabilitystack/internal/middleware"
	"github.com/enoquesousa/goobservabilitystack/internal/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	listeners, err := config.LoadListeners(os.Getenv)
	if err != nil {
		log.Fatalf("invalid listener configuration: %v", err)
	}

	// Initialise OTel TracerProvider with the configured OTLP exporter.
	// The shutdown function is called during graceful shutdown to flush pending spans.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	tp, shutdownTracing, err := tracing.New(ctx, "goobservabilitystack")
	if err != nil {
		log.Fatalf("failed to initialise tracing provider: %v", err)
	}

	taskTracer := tp.Tracer("goobservabilitystack/worker")

	reg, err := metrics.NewRegistry()
	if err != nil {
		log.Fatalf("failed to initialize metrics registry: %v", err)
	}

	worker, err := metrics.NewWorker(reg.EventsCounter(), reg.ProcessDuration(), 5*time.Second, taskTracer)
	if err != nil {
		log.Fatalf("failed to initialize metrics worker: %v", err)
	}

	go worker.Run(ctx)

	// RED middleware — instruments every app request with Rate, Errors and Duration
	// metrics and wraps each request in an OTel server span.
	httpTracer := tp.Tracer("goobservabilitystack/http")
	redMiddleware := middleware.RED(middleware.Config{
		RequestsTotal: reg.HTTPRequestsTotal(),
		Duration:      reg.HTTPDuration(),
		Tracer:        httpTracer,
	})

	// Metrics server — exposes /metrics for Prometheus scraping.
	// This server is intentionally NOT wrapped by the RED middleware to avoid
	// self-referential cardinality noise in the http_requests_total counter.
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.HandlerFor(reg.Gatherer(), promhttp.HandlerOpts{}))
	metricsServer := &http.Server{
		Addr:         listeners.MetricsAddress,
		Handler:      metricsMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Application server — all handlers are automatically observed by the RED middleware.
	appMux := http.NewServeMux()
	appMux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	appServer := &http.Server{
		Addr:         listeners.HTTPAddress,
		Handler:      redMiddleware(appMux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("metrics server listening on %s", strconv.Quote(listeners.MetricsAddress))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("metrics server error: %v", err)
		}
	}()

	go func() {
		log.Printf("app server listening on %s", strconv.Quote(listeners.HTTPAddress))
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

	// Flush and close the OTel TracerProvider last, after HTTP servers are down.
	if err := shutdownTracing(shutdownCtx); err != nil {
		log.Printf("tracing shutdown error: %v", err)
	}

	log.Println("shutdown complete")
}
