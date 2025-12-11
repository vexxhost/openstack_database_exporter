package cinder

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	cinderdb "github.com/vexxhost/openstack_database_exporter/internal/db/cinder"
)

var (
	poolFreeGbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "pool_capacity_free_gb"),
		"pool_capacity_free_gb",
		[]string{
			"name",
			"volume_backend_name",
			"vendor_name",
		},
		nil,
	)

	poolTotalGbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "pool_capacity_total_gb"),
		"pool_capacity_total_gb",
		[]string{
			"name",
			"volume_backend_name",
			"vendor_name",
		},
		nil,
	)
)

type PoolsCollector struct {
	db      *sql.DB
	queries *cinderdb.Queries
	logger  *slog.Logger
}

func NewPoolsCollector(db *sql.DB, logger *slog.Logger) *PoolsCollector {
	return &PoolsCollector{
		db:      db,
		queries: cinderdb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "pools",
		),
	}
}

func (c *PoolsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- poolFreeGbDesc
	ch <- poolTotalGbDesc
}

func (c *PoolsCollector) Collect(ch chan<- prometheus.Metric) {
}
