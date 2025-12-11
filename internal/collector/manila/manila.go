package manila

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db"
)

const (
	Namespace = "openstack"
	Subsystem = "sharev2"
)

func RegisterCollectors(registry *prometheus.Registry, databaseURL string, logger *slog.Logger) error {
	if databaseURL == "" {
		logger.Info("Collector not loaded", "service", "manila", "reason", "database URL not configured")
		return nil
	}

	conn, err := db.Connect(databaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "service", "manila", "error", err)
		return err
	}

	registry.MustRegister(NewSharesCollector(conn, logger))

	logger.Info("Registered collectors", "service", "manila")
	return nil
}
