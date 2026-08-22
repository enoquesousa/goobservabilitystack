package metrics_test

import (
	"testing"

	"github.com/enoquesousa/goobservabilitystack/internal/metrics"
)

func TestNewRegistry(t *testing.T) {
	reg, err := metrics.NewRegistry()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if reg.Gatherer() == nil {
		t.Errorf("expected gatherer not to be nil")
	}

	if reg.EventsCounter() == nil {
		t.Errorf("expected events counter not to be nil")
	}

	metricsFamilies, err := reg.Gatherer().Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	found := false
	for _, mf := range metricsFamilies {
		if mf.GetName() == "app_events_total" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("app_events_total not found in registry")
	}
}
