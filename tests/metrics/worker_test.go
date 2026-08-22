package metrics_test

import (
	"context"
	"testing"
	"time"

	"github.com/enoquesousa/goobservabilitystack/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestWorker(t *testing.T) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "test",
	})

	duration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "test_duration",
		Help:    "test",
		Buckets: prometheus.DefBuckets,
	})

	interval := 50 * time.Millisecond
	// nil tracer is accepted — Worker falls back to a no-op TracerProvider.
	worker, err := metrics.NewWorker(counter, duration, interval, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go worker.Run(ctx)

	// Wait up to 300ms per iteration + interval, giving plenty of time for 2 loops to finish.
	time.Sleep(800 * time.Millisecond)
	cancel() // stop worker

	// Expandimos esse tempo para garantir que a goroutine executando simulateProcess()
	// conclua o seu time.Sleep(50..250ms) pendente e saia do loop finalizando a thread,
	// antes que o teste termine e o Windows tente deletar o executável temporário.
	time.Sleep(350 * time.Millisecond)

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

	duration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "test_duration2",
		Help:    "test",
		Buckets: prometheus.DefBuckets,
	})

	// nil tracer is valid — not an error condition.
	_, err := metrics.NewWorker(nil, duration, 50*time.Millisecond, nil)
	if err == nil {
		t.Error("expected error with nil counter")
	}

	_, err = metrics.NewWorker(counter, nil, 50*time.Millisecond, nil)
	if err == nil {
		t.Error("expected error with nil duration")
	}

	_, err = metrics.NewWorker(counter, duration, 0, nil)
	if err == nil {
		t.Error("expected error with 0 interval")
	}

	_, err = metrics.NewWorker(counter, duration, -1*time.Second, nil)
	if err == nil {
		t.Error("expected error with negative interval")
	}
}
