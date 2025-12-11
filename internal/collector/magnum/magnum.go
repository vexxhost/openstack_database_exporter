package magnum

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	Subsystem = "container_infra"
)

func RegisterCollectors(registry *prometheus.Registry, db *sql.DB, logger *slog.Logger) {
	registry.MustRegister(NewClustersCollector(db, logger))
	registry.MustRegister(NewMastersCollector(db, logger))
	registry.MustRegister(NewNodesCollector(db, logger))
}
