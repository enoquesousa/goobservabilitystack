package metrics

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Worker increments a Prometheus Counter on a fixed interval and emits an OTel
// span for each task execution. Use [NewWorker] to construct a valid instance.
type Worker struct {
	counter  prometheus.Counter
	duration prometheus.Histogram
	interval time.Duration
	tracer   trace.Tracer
	taskSeq  atomic.Uint64
}

// NewWorker creates a [Worker] that increments the given counter every interval
// and instruments each task with an OTel span via the provided tracer.
// Pass [trace.NewNoopTracerProvider().Tracer("")] to disable tracing.
// Returns an error if counter or duration is nil, or interval is not positive.
func NewWorker(
	counter prometheus.Counter,
	duration prometheus.Histogram,
	interval time.Duration,
	tracer trace.Tracer,
) (*Worker, error) {
	if counter == nil {
		return nil, errors.New("metrics: worker counter must not be nil")
	}
	if duration == nil {
		return nil, errors.New("metrics: worker duration must not be nil")
	}
	if interval <= 0 {
		return nil, errors.New("metrics: worker interval must be positive")
	}
	if tracer == nil {
		tracer = trace.NewNoopTracerProvider().Tracer("")
	}
	return &Worker{
		counter:  counter,
		duration: duration,
		interval: interval,
		tracer:   tracer,
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
			w.simulateProcess(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// simulateProcess runs a simulated task, records Prometheus metrics, and
// emits an OTel span containing task.id, task.status, and task.duration_ms.
func (w *Worker) simulateProcess(ctx context.Context) {
	taskID := fmt.Sprintf("task-%04d", w.taskSeq.Add(1))

	ctx, span := w.tracer.Start(ctx, "worker.task",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	// Attach static attributes known at span start.
	span.SetAttributes(attribute.String("task.id", taskID))

	start := time.Now()

	// Simulate processing: random duration between 50 ms and 250 ms.
	sleepTime := time.Duration(50+rand.Intn(200)) * time.Millisecond
	time.Sleep(sleepTime)

	elapsed := time.Since(start)

	w.counter.Inc()
	w.duration.Observe(elapsed.Seconds())

	durationMS := elapsed.Milliseconds()

	// Attach result attributes after the work is done.
	span.SetAttributes(
		attribute.String("task.status", "success"),
		attribute.Int64("task.duration_ms", durationMS),
	)
	span.SetStatus(codes.Ok, "task completed successfully")
}
