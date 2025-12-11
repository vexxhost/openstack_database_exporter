package cinder

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	Subsystem = "cinder"
)

func RegisterCollectors(registry *prometheus.Registry, db *sql.DB, logger *slog.Logger) {
	registry.MustRegister(NewAgentsCollector(db, logger))
	registry.MustRegister(NewLimitsCollector(db, logger))
	registry.MustRegister(NewPoolsCollector(db, logger))
	registry.MustRegister(NewSnapshotsCollector(db, logger))
	registry.MustRegister(NewVolumesCollector(db, logger))
}