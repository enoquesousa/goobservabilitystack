package tracing_test

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
)

func TestExportUsesConfiguredTransport(t *testing.T) {
	pki := newTracePKI(t)
	for _, test := range []struct {
		name   string
		secure bool
		mutual bool
	}{
		{name: "server TLS", secure: true},
		{name: "mutual TLS", secure: true, mutual: true},
		{name: "explicit plaintext"},
	} {
		t.Run(test.name, func(t *testing.T) {
			cleanTraceEnvironment(t)
			var config *tls.Config
			scheme := "http://"
			if test.secure {
				config, scheme = pki.collectorTLS(test.mutual), "https://"
				setTraceEnvironment(t, map[string]string{"CERTIFICATE": pki.authority.path, "HEADERS": "authorization=Bearer%20synthetic-public-token"})
			}
			if test.mutual {
				setTraceEnvironment(t, map[string]string{"CLIENT_CERTIFICATE": pki.client.certPath, "CLIENT_KEY": pki.client.keyPath})
			}
			collector := startTraceCollector(t, config)
			setTraceEnvironment(t, map[string]string{"ENDPOINT": scheme + collector.address})
			if err := exportTestSpan(t); err != nil {
				t.Fatalf("span export failed: %v", err)
			}
			checkTransportObservation(t, receivedTrace(t, collector), test.secure, test.mutual)
		})
	}
}

func checkTransportObservation(t *testing.T, observation traceObservation, secure, mutual bool) {
	t.Helper()
	checkExportedSpan(t, observation)
	if observation.secure != secure || (observation.client != "") != mutual {
		t.Fatal("collector received the span through an unexpected transport identity")
	}
	if secure {
		values := observation.headers.Get("authorization")
		if len(values) != 1 || values[0] != "Bearer synthetic-public-token" {
			t.Fatal("TLS export did not preserve configured authentication metadata")
		}
	}
}

func TestTraceSpecificSettingsOverrideValidGenericSettings(t *testing.T) {
	cleanTraceEnvironment(t)
	pki, other := newTracePKI(t), newTracePKI(t)
	collector := startTraceCollector(t, pki.collectorTLS(true))
	unused := startTraceCollector(t, nil)
	setTraceEnvironment(t, map[string]string{
		"ENDPOINT": "http://" + unused.address, "INSECURE": "true",
		"CERTIFICATE": other.authority.path, "CLIENT_CERTIFICATE": other.client.certPath, "CLIENT_KEY": other.client.keyPath,
		"HEADERS":         "authorization=Bearer%20generic-token",
		"TRACES_ENDPOINT": "https://" + collector.address, "TRACES_INSECURE": "false",
		"TRACES_CERTIFICATE": pki.authority.path, "TRACES_CLIENT_CERTIFICATE": pki.client.certPath, "TRACES_CLIENT_KEY": pki.client.keyPath,
		"TRACES_HEADERS": "authorization=Bearer%20synthetic-public-token",
	})
	if err := exportTestSpan(t); err != nil {
		t.Fatalf("trace-specific credentials were not honored: %v", err)
	}
	checkTransportObservation(t, receivedTrace(t, collector), true, true)
	if unused.calls.Load() != 0 {
		t.Fatal("exporter contacted the overridden plaintext collector")
	}
}

func TestExportRejectsInvalidTLSWithoutPlaintextFallback(t *testing.T) {
	pki, other := newTracePKI(t), newTracePKI(t)
	wrongName := issueTraceIdentity(t, pki.authority, &x509.Certificate{
		Subject:     pkix.Name{CommonName: "wrong-server-name"},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, DNSNames: []string{"wrong.example"},
	})
	wrongSAN := &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{wrongName.certificate}}
	for _, test := range []struct {
		name   string
		server *tls.Config
		values map[string]string
	}{
		{name: "untrusted server", server: pki.collectorTLS(false), values: map[string]string{"CERTIFICATE": other.authority.path}},
		{name: "wrong server SAN", server: wrongSAN, values: map[string]string{"CERTIFICATE": pki.authority.path}},
		{name: "TLS to plaintext collector"},
		{name: "missing client certificate", server: pki.collectorTLS(true), values: map[string]string{"CERTIFICATE": pki.authority.path}},
		{name: "untrusted client", server: pki.collectorTLS(true), values: map[string]string{
			"CERTIFICATE": pki.authority.path, "CLIENT_CERTIFICATE": other.client.certPath, "CLIENT_KEY": other.client.keyPath,
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			cleanTraceEnvironment(t)
			collector := startTraceCollector(t, test.server)
			setTraceEnvironment(t, test.values)
			setTraceEnvironment(t, map[string]string{"ENDPOINT": "https://" + collector.address})
			if err := exportTestSpan(t); err == nil {
				t.Fatal("invalid TLS unexpectedly exported a span")
			}
			if collector.calls.Load() != 0 {
				t.Fatal("a rejected TLS client reached the OTLP export handler")
			}
		})
	}
}
