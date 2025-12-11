package octavia

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	Subsystem = "loadbalancer"
)

func RegisterCollectors(registry *prometheus.Registry, db *sql.DB, logger *slog.Logger) {
	registry.MustRegister(NewAmphoraCollector(db, logger))
	registry.MustRegister(NewLoadBalancerCollector(db, logger))
	registry.MustRegister(NewPoolCollector(db, logger))
}
