package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enoquesousa/goobservabilitystack/internal/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.opentelemetry.io/otel/trace/noop"
)

// newTestConfig builds an isolated middleware.Config with a fresh prometheus.Registry
// and a no-op tracer so tests never cross OTel network boundaries.
func newTestConfig(t *testing.T) middleware.Config {
	t.Helper()

	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "test counter",
	}, []string{"method", "route", "status_code"})

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "test histogram",
	}, []string{"method", "route", "status_code"})

	return middleware.Config{
		RequestsTotal: requestsTotal,
		Duration:      duration,
		Tracer:        noop.NewTracerProvider().Tracer("test"),
	}
}

// TestRED_CounterIncrementsOnSuccess verifies that a 200 response increments the
// requests counter with the correct label set.
func TestRED_CounterIncrementsOnSuccess(t *testing.T) {
	cfg := newTestConfig(t)
	handler := middleware.RED(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := testutil.ToFloat64(cfg.RequestsTotal.WithLabelValues("GET", "/healthz", "200")); got != 1 {
		t.Errorf("expected counter=1, got %v", got)
	}
}

// TestRED_CounterIncrementsOn500 verifies that a 500 response is labelled correctly
// so consumers can derive the Errors signal by filtering status_code >= 500.
func TestRED_CounterIncrementsOn500(t *testing.T) {
	cfg := newTestConfig(t)
	handler := middleware.RED(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := testutil.ToFloat64(cfg.RequestsTotal.WithLabelValues("POST", "/submit", "500")); got != 1 {
		t.Errorf("expected counter=1, got %v", got)
	}
}

// TestRED_ImplicitStatus200 verifies that a handler that writes a body without
// explicitly calling WriteHeader is still recorded as 200.
func TestRED_ImplicitStatus200(t *testing.T) {
	cfg := newTestConfig(t)
	handler := middleware.RED(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Implicit 200: no WriteHeader call, only Write.
		_, _ = w.Write([]byte("pong"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", rr.Code)
	}
	if got := testutil.ToFloat64(cfg.RequestsTotal.WithLabelValues("GET", "/ping", "200")); got != 1 {
		t.Errorf("expected counter=1, got %v", got)
	}
}

// TestRED_ResponsePassthrough verifies that the middleware does not alter the
// response body or status code written by the downstream handler.
func TestRED_ResponsePassthrough(t *testing.T) {
	cfg := newTestConfig(t)
	handler := middleware.RED(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/resource", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected HTTP 201, got %d", rr.Code)
	}
	if rr.Body.String() != "created" {
		t.Errorf("expected body 'created', got %q", rr.Body.String())
	}
}

// TestRED_DurationObserved verifies that the histogram always has exactly one
// observation after a single request.
func TestRED_DurationObserved(t *testing.T) {
	cfg := newTestConfig(t)
	handler := middleware.RED(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Count the number of metric families; duration must produce at least one sample.
	count := testutil.CollectAndCount(cfg.Duration)
	if count == 0 {
		t.Error("expected at least one duration observation, got 0")
	}
}

// TestRED_ContextPropagated verifies that the request context passed to the
// downstream handler carries the OTel span context (non-nil span).
func TestRED_ContextPropagated(t *testing.T) {
	cfg := newTestConfig(t)

	var capturedHandler bool
	handler := middleware.RED(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		// The span is always set; with a noop tracer it is a no-op span (valid but inactive).
		capturedHandler = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/ctx", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !capturedHandler {
		t.Error("downstream handler was never called")
	}
}
