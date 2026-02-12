package nova

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db"
)

const (
	Namespace = "openstack"
	Subsystem = "nova"
)

func RegisterCollectors(registry *prometheus.Registry, novaDatabaseURL, novaApiDatabaseURL, placementDatabaseURL string, logger *slog.Logger) {
	if novaDatabaseURL == "" || novaApiDatabaseURL == "" {
		logger.Info("Collector not loaded", "service", "nova", "reason", "database URLs not configured")
		return
	}

	novaConn, err := db.Connect(novaDatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to nova database", "service", "nova", "error", err)
		return
	}

	novaApiConn, err := db.Connect(novaApiDatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to nova_api database", "service", "nova", "error", err)
		return
	}

	var placementConn *sql.DB
	if placementDatabaseURL != "" {
		placementConn, err = db.Connect(placementDatabaseURL)
		if err != nil {
			logger.Warn("Failed to connect to placement database", "service", "nova", "error", err)
			// Continue without placement - some metrics may not be available
		}
	}

	registry.MustRegister(NewComputeCollector(novaConn, novaApiConn, placementConn, logger))

	logger.Info("Registered collectors", "service", "nova")
}
