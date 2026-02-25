package collector

import (
	"io"
	"log/slog"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewRegistry_AllEmpty(t *testing.T) {
	// When all URLs are empty, no collectors should be registered.
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	cfg := Config{}
	reg := NewRegistry(cfg, logger)

	// Gather should return no metric families (no collectors registered)
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mfs) != 0 {
		t.Fatalf("expected 0 metric families with empty config, got %d", len(mfs))
	}
}

func TestNewRegistry_ReturnsValidRegistry(t *testing.T) {
	// Verify the returned registry is a valid prometheus.Registry
	// (even with empty config, it should be usable)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	cfg := Config{}
	reg := NewRegistry(cfg, logger)

	if reg == nil {
		t.Fatal("expected non-nil registry")
	}

	// Should be a *prometheus.Registry that can be gathered from
	var _ prometheus.Gatherer = reg
	var _ prometheus.Registerer = reg
}
