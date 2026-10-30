// Package tracing provides the OpenTelemetry TracerProvider configuration
// for the goobservabilitystack service. It wires the SDK with a gRPC collector
// exporter (OTLP) to send spans to a backend like Jaeger.
package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// New constructs an OTel TracerProvider backed by an OTLP gRPC exporter.
// The provider samples 100% of spans and attaches a resource identifying the service.
// It uses the standard OTEL_EXPORTER_OTLP_* environment variables, with the
// TRACES_* variants taking precedence. The SDK defaults to https://localhost:4317.
// TLS and optional client certificates are configured by the SDK, without
// transport overrides. Plaintext requires an explicit, credential-free opt-in.
// Invalid transport settings fail before exporter creation; collector trust and
// availability are checked lazily when spans are exported.
//
// It returns:
//   - the configured TracerProvider
//   - a shutdown function that must be called on graceful shutdown to flush pending spans
//   - any initialisation error
func New(ctx context.Context, serviceName string) (trace.TracerProvider, func(context.Context) error, error) {
	if err := validateEnvironment(); err != nil {
		return nil, nil, err
	}

	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("tracing: create otlp grpc exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("0.1.0"),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("tracing: create resource: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
	)

	return provider, provider.Shutdown, nil
}
