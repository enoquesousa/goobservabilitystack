// Package metrics provides the custom Prometheus registry and all metric
// definitions for the goobservabilitystack service.
package metrics

import (
	"fmt"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
)

// Registry holds a custom Prometheus registry and the metrics registered to it.
// Use [NewRegistry] to construct a valid instance.
type Registry struct {
	reg              *prometheus.Registry
	eventsTotal      prometheus.Counter
	processDuration  prometheus.Histogram
	httpRequestsTotal *prometheus.CounterVec
	httpDuration      *prometheus.HistogramVec
}

// NewRegistry creates a new [Registry] with a custom prometheus.Registry and
// registers all application metrics to it.
// It returns an error if any metric registration fails.
func NewRegistry() (*Registry, error) {
	reg := prometheus.NewRegistry()

	eventsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "app_events_total",
		Help: "Total number of application events emitted by the background worker.",
	})
	if err := reg.Register(eventsTotal); err != nil {
		return nil, fmt.Errorf("metrics: register app_events_total: %w", err)
	}

	processDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "process_duration_seconds",
		Help:    "Duration of simulated processing in seconds.",
		Buckets: prometheus.DefBuckets,
	})
	if err := reg.Register(processDuration); err != nil {
		return nil, fmt.Errorf("metrics: register process_duration_seconds: %w", err)
	}

	activeGoroutines := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "active_goroutines",
		Help: "Current number of active goroutines in the system.",
	}, func() float64 {
		return float64(runtime.NumGoroutine())
	})
	if err := reg.Register(activeGoroutines); err != nil {
		return nil, fmt.Errorf("metrics: register active_goroutines: %w", err)
	}

	// RED pattern metrics for the HTTP layer.
	// Labels: method (HTTP verb), route (matched pattern), status_code (string).
	httpRequestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests processed, partitioned by method, route and status code.",
	}, []string{"method", "route", "status_code"})
	if err := reg.Register(httpRequestsTotal); err != nil {
		return nil, fmt.Errorf("metrics: register http_requests_total: %w", err)
	}

	httpDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds, partitioned by method, route and status code.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route", "status_code"})
	if err := reg.Register(httpDuration); err != nil {
		return nil, fmt.Errorf("metrics: register http_request_duration_seconds: %w", err)
	}

	return &Registry{
		reg:               reg,
		eventsTotal:       eventsTotal,
		processDuration:   processDuration,
		httpRequestsTotal: httpRequestsTotal,
		httpDuration:      httpDuration,
	}, nil
}

// Gatherer returns the prometheus.Gatherer interface of the underlying registry,
// suitable for use with promhttp.HandlerFor without exposing the Registerer surface.
func (r *Registry) Gatherer() prometheus.Gatherer {
	return r.reg
}

// EventsCounter returns the app_events_total Counter for use by a [Worker].
func (r *Registry) EventsCounter() prometheus.Counter {
	return r.eventsTotal
}

// ProcessDuration returns the process_duration_seconds Histogram for use by a [Worker].
func (r *Registry) ProcessDuration() prometheus.Histogram {
	return r.processDuration
}

// HTTPRequestsTotal returns the http_requests_total CounterVec for use by the RED middleware.
// Labels: method, route, status_code.
func (r *Registry) HTTPRequestsTotal() *prometheus.CounterVec {
	return r.httpRequestsTotal
}

// HTTPDuration returns the http_request_duration_seconds HistogramVec for use by the RED middleware.
// Labels: method, route, status_code.
func (r *Registry) HTTPDuration() *prometheus.HistogramVec {
	return r.httpDuration
}
