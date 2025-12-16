package ironic

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db"
)

// Namespace and subsystem constants
const (
	Namespace = "openstack"
	Subsystem = "ironic"
)

// RegisterCollectors registers all Ironic collectors with the given database and logger
func RegisterCollectors(registry *prometheus.Registry, databaseURL string, logger *slog.Logger) {
	if databaseURL == "" {
		logger.Info("Collector not loaded", "service", "ironic", "reason", "database URL not configured")
		return
	}

	conn, err := db.Connect(databaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "service", "ironic", "error", err)
		return
	}

	registry.MustRegister(NewBaremetalCollector(conn, logger))

	logger.Info("Registered collectors", "service", "ironic")
}
