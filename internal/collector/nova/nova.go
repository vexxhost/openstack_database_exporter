package nova

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/project"
	"github.com/vexxhost/openstack_database_exporter/internal/db"
	placementdb "github.com/vexxhost/openstack_database_exporter/internal/db/placement"
	"github.com/vexxhost/openstack_database_exporter/internal/util"
)

const (
	Namespace = "openstack"
	Subsystem = "nova"
)

func RegisterCollectors(registry *prometheus.Registry, novaDatabaseURL, novaApiDatabaseURL, placementDatabaseURL string, projectResolver *project.Resolver, logger *slog.Logger) {
	if novaDatabaseURL == "" || novaApiDatabaseURL == "" {
		logger.Info("Collector not loaded", "service", "nova", "reason", "database URLs not configured")
		return
	}

	novaConn, err := db.Connect(novaDatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to nova database", "service", "nova", "error", err)
		registry.MustRegister(util.NewDownCollector(Namespace, Subsystem))
		return
	}

	novaApiConn, err := db.Connect(novaApiDatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to nova_api database", "service", "nova", "error", err)
		registry.MustRegister(util.NewDownCollector(Namespace, Subsystem))
		return
	}

	var placementQueries *placementdb.Queries
	if placementDatabaseURL != "" {
		placementConn, err := db.Connect(placementDatabaseURL)
		if err != nil {
			logger.Warn("Failed to connect to placement database for Nova limits, limits_*_used metrics will be 0", "error", err)
		} else {
			placementQueries = placementdb.New(placementConn)
		}
	} else {
		logger.Warn("Placement database URL not configured, Nova limits_*_used metrics will be 0")
	}

	registry.MustRegister(NewComputeCollector(novaConn, novaApiConn, placementQueries, projectResolver, logger))

	logger.Info("Registered collectors", "service", "nova")
}
