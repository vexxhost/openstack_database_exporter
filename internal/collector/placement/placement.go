package placement

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db"
)

const (
	Namespace = "openstack"
	Subsystem = "placement"
)

func RegisterCollectors(registry *prometheus.Registry, databaseURL string, logger *slog.Logger) error {
	if databaseURL == "" {
		logger.Info("Collector not loaded", "service", "placement", "reason", "database URL not configured")
		return nil
	}

	conn, err := db.Connect(databaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "service", "placement", "error", err)
		return err
	}

	registry.MustRegister(NewResourcesCollector(conn, logger))

	logger.Info("Registered collectors", "service", "placement")
	return nil
}
