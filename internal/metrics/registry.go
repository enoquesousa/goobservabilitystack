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
	reg             *prometheus.Registry
	eventsTotal     prometheus.Counter
	processDuration prometheus.Histogram
}

// NewRegistry creates a new [Registry] with a custom prometheus.Registry and
// registers the app_events_total counter to it.
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

	return &Registry{
		reg:             reg,
		eventsTotal:     eventsTotal,
		processDuration: processDuration,
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
