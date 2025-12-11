package glance

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db"
)

const (
	Namespace = "openstack"
	Subsystem = "glance"
)

func RegisterCollectors(registry *prometheus.Registry, databaseURL string, logger *slog.Logger) {
	if databaseURL == "" {
		logger.Info("Collector not loaded", "service", "glance", "reason", "database URL not configured")
		return
	}

	conn, err := db.Connect(databaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "service", "glance", "error", err)
		return
	}

	registry.MustRegister(NewImagesCollector(conn, logger))

	logger.Info("Registered collectors", "service", "glance")
}
