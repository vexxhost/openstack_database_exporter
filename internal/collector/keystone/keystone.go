package keystone

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db"
)

const (
	Namespace = "openstack"
	Subsystem = "identity"
)

func RegisterCollectors(registry *prometheus.Registry, databaseURL string, logger *slog.Logger) error {
	if databaseURL == "" {
		logger.Info("Collector not loaded", "service", "keystone", "reason", "database URL not configured")
		return nil
	}

	conn, err := db.Connect(databaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "service", "keystone", "error", err)
		return err
	}

	registry.MustRegister(NewIdentityCollector(conn, logger))

	logger.Info("Registered collectors", "service", "keystone")
	return nil
}
