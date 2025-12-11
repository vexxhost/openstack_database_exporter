package magnum

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db"
)

const (
	Namespace = "openstack"
	Subsystem = "container_infra"
)

func RegisterCollectors(registry *prometheus.Registry, databaseURL string, logger *slog.Logger) error {
	if databaseURL == "" {
		logger.Info("Collector not loaded", "service", "magnum", "reason", "database URL not configured")
		return nil
	}

	conn, err := db.Connect(databaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "service", "magnum", "error", err)
		return err
	}

	registry.MustRegister(NewClustersCollector(conn, logger))
	registry.MustRegister(NewMastersCollector(conn, logger))
	registry.MustRegister(NewNodesCollector(conn, logger))

	logger.Info("Registered collectors", "service", "magnum")
	return nil
}
