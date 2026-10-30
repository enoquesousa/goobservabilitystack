// Package config validates the application's listener configuration before startup.
package config

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

// Listeners contains the validated application and metrics bind addresses.
type Listeners struct {
	// HTTPAddress is the application listener's host:port address.
	HTTPAddress string
	// MetricsAddress is the Prometheus listener's host:port address.
	MetricsAddress string
}

// LoadListeners reads HTTP_HOST, HTTP_PORT, METRICS_HOST and METRICS_PORT via getenv.
// Unset or empty values use loopback hosts and the existing default ports.
// It validates both listeners without resolving names or opening network resources.
func LoadListeners(getenv func(string) string) (Listeners, error) {
	httpAddress, err := listenerAddress(valueOr(getenv("HTTP_HOST"), "127.0.0.1"), valueOr(getenv("HTTP_PORT"), "8080"))
	if err != nil {
		return Listeners{}, fmt.Errorf("HTTP listener: %w", err)
	}
	metricsAddress, err := listenerAddress(valueOr(getenv("METRICS_HOST"), "127.0.0.1"), valueOr(getenv("METRICS_PORT"), "9091"))
	if err != nil {
		return Listeners{}, fmt.Errorf("metrics listener: %w", err)
	}
	return Listeners{HTTPAddress: httpAddress, MetricsAddress: metricsAddress}, nil
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func listenerAddress(host, port string) (string, error) {
	if err := validateHost(host); err != nil {
		return "", err
	}
	if err := validatePort(port); err != nil {
		return "", err
	}
	return net.JoinHostPort(host, port), nil
}

func validateHost(host string) error {
	if ip, err := netip.ParseAddr(host); err == nil {
		if ip.Zone() != "" {
			return errors.New("host must not include an IP zone")
		}
		return nil
	}
	if !validHostname(host) {
		return errors.New("host must be an IP literal or DNS hostname without a port or scheme")
	}
	return nil
}

func validatePort(port string) error {
	for _, digit := range port {
		if digit < '0' || digit > '9' {
			return errors.New("port must be a decimal number from 0 to 65535")
		}
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 0 || number > 65535 {
		return errors.New("port must be a decimal number from 0 to 65535")
	}
	return nil
}

func validHostname(host string) bool {
	host = strings.TrimSuffix(host, ".")
	if len(host) == 0 || len(host) > 253 {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if !validHostLabel(label) {
			return false
		}
	}
	return true
}

func validHostLabel(label string) bool {
	if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
		return false
	}
	for _, character := range label {
		if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') && character != '-' {
			return false
		}
	}
	return true
}
