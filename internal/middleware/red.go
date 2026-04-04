// Package middleware provides HTTP middleware for the goobservabilitystack service.
package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds the injectable dependencies required by the RED middleware.
// All fields must be non-nil. Pass [noop.NewTracerProvider().Tracer("")] to
// disable tracing without changing the middleware configuration path.
type Config struct {
	// RequestsTotal is incremented on every request, labelled by method, route and status_code.
	RequestsTotal *prometheus.CounterVec
	// Duration records the latency of each request in seconds, with the same label set.
	Duration *prometheus.HistogramVec
	// Tracer is used to open one OTel span per inbound HTTP request.
	Tracer trace.Tracer
}

// RED returns an HTTP middleware that implements the RED observability pattern:
//
//   - Rate    — http_requests_total incremented on every request
//   - Errors  — same counter filtered by status_code >= 500
//   - Duration — http_request_duration_seconds observed per request
//
// Additionally, each request is wrapped in an OTel server span. The W3C
// traceparent and tracestate headers are injected into the HTTP response so
// downstream consumers can correlate their own spans with this service's trace.
//
// The middleware is transparent to any [http.Handler] registered on the mux.
func RED(cfg Config) func(http.Handler) http.Handler {
	// Composite propagator handles both trace context and baggage via W3C headers.
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Extract incoming W3C trace context from request headers so this span
			// becomes a child of any upstream trace.
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Initial span name uses the raw path. We will update it to the matched
			// route pattern (if any) after ServeMux performs routing.
			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)

			// Open a server span named after the HTTP method and route pattern.
			ctx, span := cfg.Tracer.Start(ctx,
				spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					semconv.HTTPRequestMethodKey.String(r.Method),
					semconv.URLPath(r.URL.Path),
					semconv.ServerAddress(r.Host),
				),
			)
			defer span.End()

			// Inject the active trace context into the response headers so callers
			// can correlate downstream spans. This must happen before next.ServeHTTP
			// writes any headers to avoid "superfluous call to WriteHeader" warnings.
			propagator.Inject(ctx, propagation.HeaderCarrier(w.Header()))

			// Wrap the ResponseWriter to capture the status code written by the handler.
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			// Serve the request with the enriched context.
			next.ServeHTTP(rec, r.WithContext(ctx))

			// --- Post-handler instrumentation ---
			elapsed := time.Since(start)
			statusStr := strconv.Itoa(rec.Status())

			// Capture the matched route pattern *after* the handler runs, as
			// http.ServeMux populates r.Pattern during routing.
			route := r.Pattern
			if route == "" {
				route = r.URL.Path
			}

			// Clean the route label: if it starts with "METHOD ", remove the prefix
			// to avoid redundancy with the "method" label.
			methodPrefix := r.Method + " "
			if len(route) > len(methodPrefix) && route[:len(methodPrefix)] == methodPrefix {
				route = route[len(methodPrefix):]
			}

			cfg.RequestsTotal.WithLabelValues(r.Method, route, statusStr).Inc()
			cfg.Duration.WithLabelValues(r.Method, route, statusStr).Observe(elapsed.Seconds())

			// Finalise the span with the response status code and correct route.
			span.SetName(fmt.Sprintf("%s %s", r.Method, route))
			span.SetAttributes(
				semconv.HTTPResponseStatusCode(rec.Status()),
				attribute.String("http.route", route),
			)

			if rec.Status() >= http.StatusInternalServerError {
				span.SetStatus(codes.Error, http.StatusText(rec.Status()))
			} else {
				span.SetStatus(codes.Ok, "")
			}
		})
	}
}

// statusRecorder is a minimal [http.ResponseWriter] wrapper that captures the HTTP
// status code written by a downstream handler. It overrides both WriteHeader and
// Write so that implicit 200 responses (handler writes body without calling
// WriteHeader explicitly) are also recorded correctly.
type statusRecorder struct {
	http.ResponseWriter
	status  int
	written bool
}

// WriteHeader captures the status code on the first call and forwards it downstream.
// Subsequent calls are silently ignored to match stdlib semantics.
func (r *statusRecorder) WriteHeader(code int) {
	if !r.written {
		r.status = code
		r.written = true
	}
	r.ResponseWriter.WriteHeader(code)
}

// Write ensures that an implicit 200 is recorded before delegating to the
// underlying ResponseWriter, which may trigger its own implicit WriteHeader(200).
func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(b)
}

// Status returns the captured HTTP status code.
// If the handler never called WriteHeader or Write, it returns 200 (the default).
func (r *statusRecorder) Status() int { return r.status }
