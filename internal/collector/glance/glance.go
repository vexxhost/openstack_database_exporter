package glance

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	Subsystem = "glance"
)

func RegisterCollectors(registry *prometheus.Registry, db *sql.DB, logger *slog.Logger) {
	registry.MustRegister(NewImagesCollector(db, logger))
}
