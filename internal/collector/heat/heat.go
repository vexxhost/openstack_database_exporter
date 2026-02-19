package heat

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db"
	"github.com/vexxhost/openstack_database_exporter/internal/util"
)

const (
	Namespace = "openstack"
	Subsystem = "heat"
)

func RegisterCollectors(registry *prometheus.Registry, databaseURL string, logger *slog.Logger) {
	if databaseURL == "" {
		logger.Info("Collector not loaded", "service", "heat", "reason", "database URL not configured")
		return
	}

	conn, err := db.Connect(databaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "service", "heat", "error", err)
		registry.MustRegister(util.NewDownCollector(Namespace, Subsystem))
		return
	}

	registry.MustRegister(NewStacksCollector(conn, logger))

	logger.Info("Registered collectors", "service", "heat")
}
