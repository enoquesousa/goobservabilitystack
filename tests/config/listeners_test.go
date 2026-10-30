package config_test

import (
	"strings"
	"testing"

	"github.com/enoquesousa/goobservabilitystack/internal/config"
)

func TestListenersDefaultToLoopbackAndPreserveExplicitOverrides(t *testing.T) {
	for _, test := range []struct {
		name string
		env  map[string]string
		want config.Listeners
	}{
		{"unset", nil, config.Listeners{HTTPAddress: "127.0.0.1:8080", MetricsAddress: "127.0.0.1:9091"}},
		{"empty values", map[string]string{"HTTP_HOST": "", "HTTP_PORT": "", "METRICS_HOST": "", "METRICS_PORT": ""}, config.Listeners{HTTPAddress: "127.0.0.1:8080", MetricsAddress: "127.0.0.1:9091"}},
		{"existing port overrides", map[string]string{"HTTP_PORT": "8088", "METRICS_PORT": "9199"}, config.Listeners{HTTPAddress: "127.0.0.1:8088", MetricsAddress: "127.0.0.1:9199"}},
		{"IPv6", map[string]string{"HTTP_HOST": "::1", "METRICS_HOST": "2001:db8::1"}, config.Listeners{HTTPAddress: "[::1]:8080", MetricsAddress: "[2001:db8::1]:9091"}},
		{"explicit wildcard", map[string]string{"HTTP_HOST": "0.0.0.0", "METRICS_HOST": "::"}, config.Listeners{HTTPAddress: "0.0.0.0:8080", MetricsAddress: "[::]:9091"}},
		{"explicit remote IPv4", map[string]string{"HTTP_HOST": "192.0.2.10", "METRICS_HOST": "192.0.2.11"}, config.Listeners{HTTPAddress: "192.0.2.10:8080", MetricsAddress: "192.0.2.11:9091"}},
		{"explicit DNS", map[string]string{"HTTP_HOST": "localhost", "METRICS_HOST": "metrics.internal."}, config.Listeners{HTTPAddress: "localhost:8080", MetricsAddress: "metrics.internal.:9091"}},
		{"ephemeral ports", map[string]string{"HTTP_PORT": "0", "METRICS_PORT": "0"}, config.Listeners{HTTPAddress: "127.0.0.1:0", MetricsAddress: "127.0.0.1:0"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			actual, err := config.LoadListeners(func(key string) string { return test.env[key] })
			if err != nil || actual != test.want {
				t.Fatalf("listeners = %+v, error = %v; want %+v", actual, err, test.want)
			}
		})
	}
}

func TestListenersRejectInvalidHostsAndPorts(t *testing.T) {
	for _, field := range []string{"HTTP_HOST", "METRICS_HOST"} {
		for _, value := range []string{
			"http://127.0.0.1", "127.0.0.1:8080", "[::1]", "::1%lo", "fe80::1%eth0",
			"127.0.0.1 ", " localhost", "user@localhost", "host/path", "host..example", "-host", "host-",
			"host_name", "éxample.test", strings.Repeat("a", 64) + ".example", strings.Repeat("a.", 128),
		} {
			t.Run(field+"/"+value, func(t *testing.T) { assertInvalidListener(t, field, value) })
		}
	}
	for _, field := range []string{"HTTP_PORT", "METRICS_PORT"} {
		for _, value := range []string{"-1", "+8080", "65536", "9999999999999999999999", "http", " 8080", "8080 ", "８０８０", "80:80"} {
			t.Run(field+"/"+value, func(t *testing.T) { assertInvalidListener(t, field, value) })
		}
	}
}

func TestListenerErrorsDoNotReflectMaliciousConfiguration(t *testing.T) {
	const marker = "FORGED_CONFIGURATION_LINE"
	for _, field := range []string{"HTTP_HOST", "HTTP_PORT", "METRICS_HOST", "METRICS_PORT"} {
		t.Run(field, func(t *testing.T) {
			_, err := config.LoadListeners(func(key string) string {
				if key == field {
					return "127.0.0.1\n" + marker
				}
				return ""
			})
			if err == nil || strings.Contains(err.Error(), marker) || strings.ContainsAny(err.Error(), "\r\n") {
				t.Fatalf("expected a sanitized configuration error, got %v", err)
			}
		})
	}
}

func assertInvalidListener(t *testing.T, field, value string) {
	t.Helper()
	actual, err := config.LoadListeners(func(key string) string {
		if key == field {
			return value
		}
		return ""
	})
	if err == nil {
		t.Fatalf("accepted invalid %s", field)
	}
	if actual != (config.Listeners{}) {
		t.Fatalf("returned partially validated configuration: %+v", actual)
	}
}
