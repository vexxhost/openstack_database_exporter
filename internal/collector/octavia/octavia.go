package octavia

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db"
)

const (
	Namespace = "openstack"
	Subsystem = "loadbalancer"
)

func RegisterCollectors(registry *prometheus.Registry, databaseURL string, logger *slog.Logger) {
	if databaseURL == "" {
		logger.Info("Collector not loaded", "service", "octavia", "reason", "database URL not configured")
		return
	}

	conn, err := db.Connect(databaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "service", "octavia", "error", err)
		return
	}

	registry.MustRegister(NewAmphoraCollector(conn, logger))
	registry.MustRegister(NewLoadBalancerCollector(conn, logger))
	registry.MustRegister(NewPoolCollector(conn, logger))

	logger.Info("Registered collectors", "service", "octavia")
}
