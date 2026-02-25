package cinder

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/project"
	"github.com/vexxhost/openstack_database_exporter/internal/db"
	"github.com/vexxhost/openstack_database_exporter/internal/util"
)

const (
	Namespace = "openstack"
	Subsystem = "cinder"
)

func RegisterCollectors(registry *prometheus.Registry, databaseURL string, projectResolver *project.Resolver, logger *slog.Logger) {
	if databaseURL == "" {
		logger.Info("Collector not loaded", "service", "cinder", "reason", "database URL not configured")
		return
	}

	conn, err := db.Connect(databaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "service", "cinder", "error", err)
		registry.MustRegister(util.NewDownCollector(Namespace, Subsystem))
		return
	}

	registry.MustRegister(NewAgentsCollector(conn, logger))
	registry.MustRegister(NewLimitsCollector(conn, logger, projectResolver))
	registry.MustRegister(NewSnapshotsCollector(conn, logger))
	registry.MustRegister(NewVolumesCollector(conn, logger))

	logger.Info("Registered collectors", "service", "cinder")
}
