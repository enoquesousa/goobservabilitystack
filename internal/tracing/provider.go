// Package tracing provides the OpenTelemetry TracerProvider configuration
// for the goobservabilitystack service. It wires the SDK with a stdout exporter
// so that spans are printed to the console for local demonstration.
package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// New constructs an OTel TracerProvider backed by a stdout exporter.
// The provider samples 100% of spans and attaches a resource identifying the service.
//
// It returns:
//   - the configured TracerProvider
//   - a shutdown function that must be called on graceful shutdown to flush pending spans
//   - any initialisation error
func New(ctx context.Context, serviceName string) (trace.TracerProvider, func(context.Context) error, error) {
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("tracing: create stdout exporter: %w", err)
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
		// SimpleSpanProcessor exports each span synchronously — ideal for demos
		// where you want immediate stdout output without batching delays.
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
	)

	return provider, provider.Shutdown, nil
}
