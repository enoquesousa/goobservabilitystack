package tracing_test

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/enoquesousa/goobservabilitystack/internal/tracing"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type traceObservation struct {
	request *collectortrace.ExportTraceServiceRequest
	secure  bool
	client  string
	headers metadata.MD
}

type traceCollector struct {
	collectortrace.UnimplementedTraceServiceServer
	address      string
	calls        atomic.Int32
	observations chan traceObservation
}

// Export records a real OTLP request and its authenticated transport identity.
func (collector *traceCollector) Export(ctx context.Context, request *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error) {
	observation := traceObservation{request: request}
	observation.headers, _ = metadata.FromIncomingContext(ctx)
	if remote, ok := peer.FromContext(ctx); ok {
		if info, secure := remote.AuthInfo.(credentials.TLSInfo); secure {
			observation.secure = info.State.HandshakeComplete && info.State.Version >= tls.VersionTLS12
			if len(info.State.VerifiedChains) > 0 && len(info.State.PeerCertificates) > 0 {
				observation.client = info.State.PeerCertificates[0].Subject.CommonName
			}
		}
	}
	collector.calls.Add(1)
	select {
	case collector.observations <- observation:
	default:
	}
	return &collectortrace.ExportTraceServiceResponse{}, nil
}

func startTraceCollector(t *testing.T, tlsConfig *tls.Config) *traceCollector {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var options []grpc.ServerOption
	if tlsConfig != nil {
		options = append(options, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}
	server := grpc.NewServer(options...)
	collector := &traceCollector{address: listener.Addr().String(), observations: make(chan traceObservation, 2)}
	collectortrace.RegisterTraceServiceServer(server, collector)
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		if serveErr := <-done; serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			t.Error(serveErr)
		}
	})
	return collector
}

func cleanTraceEnvironment(t *testing.T) {
	t.Helper()
	for _, prefix := range []string{"OTEL_EXPORTER_OTLP_", "OTEL_EXPORTER_OTLP_TRACES_"} {
		for _, suffix := range []string{"ENDPOINT", "INSECURE", "CERTIFICATE", "CLIENT_CERTIFICATE", "CLIENT_KEY", "HEADERS", "TIMEOUT", "COMPRESSION"} {
			t.Setenv(prefix+suffix, "")
		}
	}
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_TIMEOUT", "400")
}

func setTraceEnvironment(t *testing.T, values map[string]string) {
	t.Helper()
	for suffix, value := range values {
		t.Setenv("OTEL_EXPORTER_OTLP_"+suffix, value)
	}
}

func exportTestSpan(t *testing.T) error {
	t.Helper()
	provider, shutdown, err := tracing.New(context.Background(), "trace-transport-test")
	if err != nil {
		t.Fatalf("valid transport configuration failed at creation: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if shutdownErr := shutdown(ctx); shutdownErr != nil {
			t.Error(shutdownErr)
		}
	})
	_, span := provider.Tracer("transport-regression").Start(context.Background(), "transport-regression")
	span.End()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	sdkProvider, ok := provider.(*sdktrace.TracerProvider)
	if !ok {
		t.Fatal("expected a real SDK trace provider")
	}
	return sdkProvider.ForceFlush(ctx)
}

func receivedTrace(t *testing.T, collector *traceCollector) traceObservation {
	t.Helper()
	select {
	case observation := <-collector.observations:
		return observation
	case <-time.After(time.Second):
		t.Fatal("collector did not receive an OTLP export")
		return traceObservation{}
	}
}

func checkExportedSpan(t *testing.T, observation traceObservation) {
	t.Helper()
	resources := observation.request.GetResourceSpans()
	if len(resources) != 1 || len(resources[0].GetScopeSpans()) != 1 {
		t.Fatal("collector received an unexpected resource or scope count")
	}
	spans := resources[0].GetScopeSpans()[0].GetSpans()
	if len(spans) != 1 || spans[0].GetName() != "transport-regression" {
		t.Fatal("collector did not receive the actual test span")
	}
	for _, attribute := range resources[0].GetResource().GetAttributes() {
		if attribute.GetKey() == "service.name" && attribute.GetValue().GetStringValue() == "trace-transport-test" {
			return
		}
	}
	t.Fatal("exported span omitted the service identity")
}
