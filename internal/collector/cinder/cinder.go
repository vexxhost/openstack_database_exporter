package cinder

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db"
)

const (
	Namespace = "openstack"
	Subsystem = "cinder"
)

func RegisterCollectors(registry *prometheus.Registry, databaseURL string, logger *slog.Logger) {
	if databaseURL == "" {
		logger.Info("Collector not loaded", "service", "cinder", "reason", "database URL not configured")
		return
	}

	conn, err := db.Connect(databaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "service", "cinder", "error", err)
		return
	}

	registry.MustRegister(NewAgentsCollector(conn, logger))
	registry.MustRegister(NewLimitsCollector(conn, logger))
	registry.MustRegister(NewPoolsCollector(conn, logger))
	registry.MustRegister(NewSnapshotsCollector(conn, logger))
	registry.MustRegister(NewVolumesCollector(conn, logger))

	logger.Info("Registered collectors", "service", "cinder")
}
