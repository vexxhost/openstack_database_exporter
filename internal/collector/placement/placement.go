package placement

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	Subsystem = "placement"
)

func RegisterCollectors(registry *prometheus.Registry, db *sql.DB, logger *slog.Logger) {
	registry.MustRegister(NewResourcesCollector(db, logger))
}
