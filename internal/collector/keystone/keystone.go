package keystone

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	Subsystem = "identity"
)

func RegisterCollectors(registry *prometheus.Registry, db *sql.DB, logger *slog.Logger) {
	registry.MustRegister(NewIdentityCollector(db, logger))
}
