package metrics_test

import (
	"context"
	"fmt"
	"time"

	"github.com/enoquesousa/goobservabilitystack/internal/metrics"
	"go.opentelemetry.io/otel/trace/noop"
)

// ExampleNewRegistry demonstrates how to create a new metrics registry
// and access its counters.
func ExampleNewRegistry() {
	reg, err := metrics.NewRegistry()
	if err != nil {
		panic(err)
	}

	// Increment a counter
	reg.EventsCounter().Inc()

	fmt.Println("Registry initialized and counter incremented")
	// Output: Registry initialized and counter incremented
}

// ExampleWorker demonstrates how to create and run a background worker
// that records metrics and emits traces.
func ExampleWorker() {
	reg, _ := metrics.NewRegistry()
	tracer := noop.NewTracerProvider().Tracer("example")

	// Create a worker that ticks every 100ms
	worker, err := metrics.NewWorker(
		reg.EventsCounter(),
		reg.ProcessDuration(),
		100*time.Millisecond,
		tracer,
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	// Run the worker (it will stop when the context is cancelled)
	worker.Run(ctx)

	fmt.Println("Worker finished execution")
	// Output: Worker finished execution
}
