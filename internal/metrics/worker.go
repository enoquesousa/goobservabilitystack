package metrics

import (
	"context"
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Worker increments a Prometheus Counter on a fixed interval.
// Use [NewWorker] to construct a valid instance.
type Worker struct {
	counter  prometheus.Counter
	interval time.Duration
}

// NewWorker creates a [Worker] that increments the given counter every interval.
// It returns an error if counter is nil or interval is not positive.
func NewWorker(counter prometheus.Counter, interval time.Duration) (*Worker, error) {
	if counter == nil {
		return nil, errors.New("metrics: worker counter must not be nil")
	}
	if interval <= 0 {
		return nil, errors.New("metrics: worker interval must be positive")
	}
	return &Worker{
		counter:  counter,
		interval: interval,
	}, nil
}

// Run starts the Worker loop, blocking until ctx is cancelled.
// It is safe to call Run in a goroutine. The internal ticker is stopped
// before Run returns, preventing goroutine leaks.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.counter.Inc()
		case <-ctx.Done():
			return
		}
	}
}
