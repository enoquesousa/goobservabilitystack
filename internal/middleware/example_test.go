package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/ESousa97/goobservabilitystack/internal/metrics"
	"github.com/ESousa97/goobservabilitystack/internal/middleware"
	"go.opentelemetry.io/otel/trace/noop"
)

// ExampleRED demonstrates how to wrap an HTTP handler with the RED middleware
// to automatically record metrics and traces.
func ExampleRED() {
	reg, _ := metrics.NewRegistry()
	tracer := noop.NewTracerProvider().Tracer("example")

	// Configure the RED middleware
	redMiddleware := middleware.RED(middleware.Config{
		RequestsTotal: reg.HTTPRequestsTotal(),
		Duration:      reg.HTTPDuration(),
		Tracer:        tracer,
	})

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "Hello, Observability!")
	})

	// Wrap the handler with the middleware
	observedHandler := redMiddleware(handler)

	// Simulate a request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	observedHandler.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	// Output: Status: 200
}
