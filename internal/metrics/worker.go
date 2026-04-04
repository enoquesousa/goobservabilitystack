package metrics

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Worker increments a Prometheus Counter on a fixed interval.
// Use [NewWorker] to construct a valid instance.
type Worker struct {
	counter  prometheus.Counter
	duration prometheus.Histogram
	interval time.Duration
}

// NewWorker creates a [Worker] that increments the given counter every interval.
// It returns an error if counter is nil or interval is not positive.
func NewWorker(counter prometheus.Counter, duration prometheus.Histogram, interval time.Duration) (*Worker, error) {
	if counter == nil {
		return nil, errors.New("metrics: worker counter must not be nil")
	}
	if duration == nil {
		return nil, errors.New("metrics: worker duration must not be nil")
	}
	if interval <= 0 {
		return nil, errors.New("metrics: worker interval must be positive")
	}
	return &Worker{
		counter:  counter,
		duration: duration,
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
			w.simulateProcess()
		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) simulateProcess() {
	start := time.Now()
	
	// Simulating process duration between 50 and 250 milliseconds
	sleepTime := time.Duration(50+rand.Intn(200)) * time.Millisecond
	time.Sleep(sleepTime)

	w.counter.Inc()
	w.duration.Observe(time.Since(start).Seconds())
}
