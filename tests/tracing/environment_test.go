package tracing_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/enoquesousa/goobservabilitystack/internal/tracing"
)

// This is a public marker used to detect accidental disclosure in diagnostics.
const traceSecretMarker = "synthetic-private-setting-marker"

type invalidTraceSetting struct {
	name   string
	values map[string]string
}

func TestInvalidEnvironmentFailsWithoutSecretLogs(t *testing.T) {
	if os.Getenv("TRACE_CONFIG_LOG_PROBE") == "1" {
		runInvalidTraceSettings(t)
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^TestInvalidEnvironmentFailsWithoutSecretLogs$")
	command.Env = append(os.Environ(), "TRACE_CONFIG_LOG_PROBE=1")
	output, err := command.CombinedOutput()
	if strings.Contains(string(output), traceSecretMarker) {
		t.Fatal("invalid OTLP settings leaked a credential marker to process diagnostics")
	}
	if err != nil {
		t.Fatalf("invalid-setting subprocess failed: %v\n%s", err, output)
	}
}

func runInvalidTraceSettings(t *testing.T) {
	t.Helper()
	pki, other := newTracePKI(t), newTracePKI(t)
	badFile := filepath.Join(t.TempDir(), traceSecretMarker+".pem")
	if err := os.WriteFile(badFile, []byte(traceSecretMarker), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, test := range invalidTraceSettings(pki, other, badFile) {
		t.Run(test.name, func(t *testing.T) {
			cleanTraceEnvironment(t)
			setTraceEnvironment(t, test.values)
			provider, shutdown, err := tracing.New(context.Background(), "invalid-transport-test")
			if shutdown != nil {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				if closeErr := shutdown(ctx); closeErr != nil {
					t.Error("unexpected provider could not shut down")
				}
			}
			if err == nil || provider != nil || shutdown != nil {
				t.Fatal("invalid setting did not fail before provider construction")
			}
			if strings.Contains(err.Error(), traceSecretMarker) || strings.Contains(err.Error(), pki.client.keyPath) {
				t.Fatal("configuration error disclosed a value or private key path")
			}
		})
	}
}

func invalidTraceSettings(pki, other tracePKI, badFile string) []invalidTraceSetting {
	return []invalidTraceSetting{
		{name: "missing CA file", values: map[string]string{"CERTIFICATE": badFile + ".missing"}},
		{name: "invalid CA PEM", values: map[string]string{"CERTIFICATE": badFile}},
		{name: "invalid generic CA despite valid override", values: map[string]string{"CERTIFICATE": badFile, "TRACES_CERTIFICATE": pki.authority.path}},
		{name: "client certificate without key", values: map[string]string{"CLIENT_CERTIFICATE": pki.client.certPath}},
		{name: "client key without certificate", values: map[string]string{"CLIENT_KEY": pki.client.keyPath}},
		{name: "trace certificate without key", values: map[string]string{"TRACES_CLIENT_CERTIFICATE": pki.client.certPath}},
		{name: "invalid client PEM", values: map[string]string{"CLIENT_CERTIFICATE": pki.client.certPath, "CLIENT_KEY": badFile}},
		{name: "mismatched client key", values: map[string]string{"CLIENT_CERTIFICATE": pki.client.certPath, "CLIENT_KEY": other.client.keyPath}},
		{name: "partial generic pair despite valid override", values: map[string]string{
			"CLIENT_CERTIFICATE": pki.client.certPath, "TRACES_CLIENT_CERTIFICATE": pki.client.certPath, "TRACES_CLIENT_KEY": pki.client.keyPath,
		}},
		{name: "URL parse failure", values: map[string]string{"ENDPOINT": "https://%zz/" + traceSecretMarker}},
		{name: "URL credentials", values: map[string]string{"ENDPOINT": "https://user:" + traceSecretMarker + "@localhost:4317"}},
		{name: "URL query", values: map[string]string{"ENDPOINT": "https://localhost:4317?token=" + traceSecretMarker}},
		{name: "empty URL query", values: map[string]string{"ENDPOINT": "https://localhost:4317?"}},
		{name: "URL fragment", values: map[string]string{"ENDPOINT": "https://localhost:4317#" + traceSecretMarker}},
		{name: "missing URL scheme", values: map[string]string{"ENDPOINT": "localhost:4317"}},
		{name: "invalid boolean", values: map[string]string{"INSECURE": traceSecretMarker}},
		{name: "invalid timeout despite valid override", values: map[string]string{"TIMEOUT": traceSecretMarker, "TRACES_TIMEOUT": "400"}},
		{name: "invalid trace timeout", values: map[string]string{"TRACES_TIMEOUT": traceSecretMarker}},
		{name: "HTTPS with insecure flag", values: map[string]string{"ENDPOINT": "https://localhost:4317", "INSECURE": "true"}},
		{name: "HTTP with secure flag", values: map[string]string{"ENDPOINT": "http://localhost:4317", "INSECURE": "false"}},
		{name: "headers over plaintext", values: map[string]string{"ENDPOINT": "http://localhost:4317", "HEADERS": "authorization=" + traceSecretMarker}},
		{name: "default endpoint made insecure with headers", values: map[string]string{"INSECURE": "true", "HEADERS": "authorization=" + traceSecretMarker}},
		{name: "client material with plaintext", values: map[string]string{
			"ENDPOINT": "http://localhost:4317", "CLIENT_CERTIFICATE": pki.client.certPath, "CLIENT_KEY": pki.client.keyPath,
		}},
		{name: "CA material with plaintext", values: map[string]string{"ENDPOINT": "http://localhost:4317", "CERTIFICATE": pki.authority.path}},
		{name: "header without separator", values: map[string]string{"HEADERS": traceSecretMarker}},
		{name: "invalid header key", values: map[string]string{"HEADERS": traceSecretMarker + ":=value"}},
		{name: "Unicode key that folds to ASCII", values: map[string]string{"HEADERS": traceSecretMarker + "K=value"}},
		{name: "invalid header escape", values: map[string]string{"HEADERS": "authorization=" + traceSecretMarker + "%zz"}},
		{name: "control character in header", values: map[string]string{"HEADERS": "authorization=" + traceSecretMarker + "%0Ainside"}},
		{name: "invalid generic header despite valid override", values: map[string]string{"HEADERS": traceSecretMarker, "TRACES_HEADERS": "authorization=Bearer%20public-test"}},
	}
}
