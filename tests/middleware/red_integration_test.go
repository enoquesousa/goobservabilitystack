// Package middleware_test contains end-to-end integration tests for the RED middleware.
// They spin up a real httptest.Server, fire HTTP requests and verify that the
// Prometheus metrics are populated and the W3C traceparent header is injected.
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ESousa97/goobservabilitystack/internal/middleware"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// newRealTracer builds a real TracerProvider backed by a discarding exporter so
// the middleware generates valid trace IDs without polluting stdout during tests.
func newRealTracer(t *testing.T) trace.Tracer {
	t.Helper()

	// stdout exporter with /dev/null-equivalent: we do not assert on span output here.
	exp, err := stdouttrace.New(stdouttrace.WithWriter(&discardWriter{}))
	if err != nil {
		t.Fatalf("failed to create stdout trace exporter: %v", err)
	}

	res, err := resource.New(t.Context(),
		resource.WithAttributes(semconv.ServiceName("test")),
	)
	if err != nil {
		t.Fatalf("failed to create OTel resource: %v", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exp)),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
	)
	t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })

	return provider.Tracer("integration-test")
}

// discardWriter implements io.Writer by silently dropping all data.
type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

// newIntegrationConfig builds a complete middleware.Config for integration tests.
func newIntegrationConfig(t *testing.T) (middleware.Config, *prometheus.Registry) {
	t.Helper()

	reg := prometheus.NewRegistry()

	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
	}, []string{"method", "route", "status_code"})
	reg.MustRegister(requestsTotal)

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
	}, []string{"method", "route", "status_code"})
	reg.MustRegister(duration)

	return middleware.Config{
		RequestsTotal: requestsTotal,
		Duration:      duration,
		Tracer:        newRealTracer(t),
	}, reg
}

// gatherCounter collects all metric families from reg and returns the value of the
// counter matching metricName with the supplied label pairs.
func gatherCounter(t *testing.T, reg *prometheus.Registry, metricName string, labels map[string]string) float64 {
	t.Helper()

	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("prometheus gather error: %v", err)
	}

	for _, mf := range families {
		if mf.GetName() != metricName {
			continue
		}
		for _, m := range mf.GetMetric() {
			if labelsMatch(m.GetLabel(), labels) {
				if c := m.GetCounter(); c != nil {
					return c.GetValue()
				}
			}
		}
	}
	return 0
}

// labelsMatch returns true when every pair in want is present in got.
func labelsMatch(got []*dto.LabelPair, want map[string]string) bool {
	index := make(map[string]string, len(got))
	for _, lp := range got {
		index[lp.GetName()] = lp.GetValue()
	}
	for k, v := range want {
		if index[k] != v {
			return false
		}
	}
	return true
}

// TestIntegration_RedMetricsPopulated fires a GET /healthz through a real
// httptest.Server and asserts that http_requests_total is exactly 1.
func TestIntegration_RedMetricsPopulated(t *testing.T) {
	cfg, reg := newIntegrationConfig(t)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(middleware.RED(cfg)(mux))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/healthz") //nolint:noctx
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", resp.StatusCode)
	}

	got := gatherCounter(t, reg, "http_requests_total", map[string]string{
		"method":      "GET",
		"route":       "/healthz",
		"status_code": "200",
	})
	if got != 1 {
		t.Errorf("expected http_requests_total=1, got %v", got)
	}
}

// TestIntegration_TraceparentHeaderInjected verifies that the W3C traceparent
// header is present in the HTTP response when a real tracer is active.
func TestIntegration_TraceparentHeaderInjected(t *testing.T) {
	cfg, _ := newIntegrationConfig(t)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(middleware.RED(cfg)(mux))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/healthz") //nolint:noctx
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	tp := resp.Header.Get("Traceparent")
	if tp == "" {
		t.Fatal("expected 'traceparent' header in response, got none")
	}
	// W3C trace context format: 00-<traceId>-<spanId>-<flags>
	if !strings.HasPrefix(tp, "00-") {
		t.Errorf("traceparent header has unexpected format: %q", tp)
	}
}

// TestIntegration_ErrorCounterLabelled verifies that a 500 response is labelled
// with status_code=500 so consumers can derive the Errors signal.
func TestIntegration_ErrorCounterLabelled(t *testing.T) {
	cfg, reg := newIntegrationConfig(t)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /boom", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	srv := httptest.NewServer(middleware.RED(cfg)(mux))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/boom") //nolint:noctx
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	got := gatherCounter(t, reg, "http_requests_total", map[string]string{
		"method":      "GET",
		"route":       "/boom",
		"status_code": "500",
	})
	if got != 1 {
		t.Errorf("expected error counter=1, got %v", got)
	}
}
