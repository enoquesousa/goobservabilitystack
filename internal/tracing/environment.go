package tracing

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// The SDK logs and skips some invalid environment settings. Validate transport,
// credentials and diagnostic-bearing values before construction, including
// values later overridden by TRACES_ settings. The SDK still owns configuration
// and precedence.
func validateEnvironment() error {
	checks := []struct {
		suffix string
		check  func(string) error
	}{
		{"ENDPOINT", validateEndpoint},
		{"INSECURE", validateInsecure},
		{"CERTIFICATE", validateCertificate},
		{"HEADERS", validateHeaders},
		{"TIMEOUT", validateTimeout},
	}
	for _, prefix := range []string{"OTEL_EXPORTER_OTLP_", "OTEL_EXPORTER_OTLP_TRACES_"} {
		for _, check := range checks {
			if err := check.check(prefix + check.suffix); err != nil {
				return err
			}
		}
		if err := validateClientCertificate(prefix); err != nil {
			return err
		}
	}
	return validateTransportSelection()
}

func setting(name string) string {
	return strings.TrimSpace(os.Getenv(name))
}

func effectiveSetting(suffix string) string {
	if value := setting("OTEL_EXPORTER_OTLP_TRACES_" + suffix); value != "" {
		return value
	}
	return setting("OTEL_EXPORTER_OTLP_" + suffix)
}

func invalidSetting(name, reason string) error {
	// Only constant environment variable names and descriptions reach this error;
	// paths, endpoints, headers and credentials must never be included.
	return fmt.Errorf("tracing: %s %s", name, reason)
}

func validateEndpoint(name string) error {
	value := setting(name)
	if value == "" {
		return nil
	}
	endpoint, err := url.Parse(value)
	if err != nil {
		return invalidSetting(name, "must be a valid HTTP or HTTPS collector URL")
	}
	scheme := strings.ToLower(endpoint.Scheme)
	if (scheme != "https" && scheme != "http") || endpoint.Hostname() == "" {
		return invalidSetting(name, "must include an HTTP or HTTPS scheme and host")
	}
	if endpoint.User != nil || endpoint.RawQuery != "" || endpoint.ForceQuery || strings.Contains(value, "#") {
		return invalidSetting(name, "must not contain user information, a query or a fragment")
	}
	return nil
}

func validateInsecure(name string) error {
	value := setting(name)
	if value != "" && !strings.EqualFold(value, "true") && !strings.EqualFold(value, "false") {
		return invalidSetting(name, "must be true or false when set")
	}
	return nil
}

func validateTimeout(name string) error {
	if value := setting(name); value != "" {
		if _, err := strconv.Atoi(value); err != nil {
			return invalidSetting(name, "must contain an integer number of milliseconds")
		}
	}
	return nil
}

func validateTransportSelection() error {
	plaintext, err := plaintextSelected()
	if err != nil || !plaintext {
		return err
	}
	for _, suffix := range []string{"CERTIFICATE", "CLIENT_CERTIFICATE", "CLIENT_KEY", "HEADERS"} {
		if effectiveSetting(suffix) != "" {
			return invalidSetting("OTLP plaintext transport", "must not be combined with TLS material or export headers")
		}
	}
	return nil
}

func plaintextSelected() (bool, error) {
	insecure := effectiveSetting("INSECURE")
	endpoint := effectiveSetting("ENDPOINT")
	if endpoint == "" {
		return strings.EqualFold(insecure, "true"), nil
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return false, invalidSetting("OTLP endpoint", "must be a valid collector URL")
	}
	plaintext := strings.EqualFold(parsed.Scheme, "http")
	if insecure != "" && strings.EqualFold(insecure, "true") != plaintext {
		return false, invalidSetting("OTLP endpoint scheme and INSECURE", "must not contradict each other")
	}
	return plaintext, nil
}
