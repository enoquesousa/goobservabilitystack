package metrics_test

import (
	"context"
	"testing"
	"time"

	"github.com/ESousa97/goobservabilitystack/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestWorker(t *testing.T) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "test",
	})

	interval := 50 * time.Millisecond
	worker, err := metrics.NewWorker(counter, interval)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go worker.Run(ctx)

	// Wait for a little more than 2 intervals
	time.Sleep(120 * time.Millisecond)
	cancel() // stop worker

	// A wait interval to guarantee that any running goroutine had time to process Cancel 
	// before reading the metric
	time.Sleep(10 * time.Millisecond)

	val := testutil.ToFloat64(counter)

	if val < 2 {
		t.Errorf("expected counter >= 2, got %v", val)
	}
}

func TestNewWorker_Errors(t *testing.T) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter2",
		Help: "test",
	})

	_, err := metrics.NewWorker(nil, 50*time.Millisecond)
	if err == nil {
		t.Error("expected error with nil counter")
	}

	_, err = metrics.NewWorker(counter, 0)
	if err == nil {
		t.Error("expected error with 0 interval")
	}

	_, err = metrics.NewWorker(counter, -1*time.Second)
	if err == nil {
		t.Error("expected error with negative interval")
	}
}
