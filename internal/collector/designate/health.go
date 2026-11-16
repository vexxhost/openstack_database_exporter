package designate

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

const healthQuery = "SELECT COUNT(*) FROM zones"

var designateUpDesc = prometheus.NewDesc(
	prometheus.BuildFQName(Namespace, Subsystem, "up"),
	"up",
	nil,
	nil,
)

type HealthCollector struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewHealthCollector(db *sql.DB, logger *slog.Logger) *HealthCollector {
	return &HealthCollector{
		db: db,
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "health",
		),
	}
}

func (c *HealthCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- designateUpDesc
}

func (c *HealthCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	var result int
	if err := c.db.QueryRowContext(ctx, healthQuery).Scan(&result); err != nil {
		c.logger.Error("Failed to collect designate health", "error", err)
		ch <- prometheus.MustNewConstMetric(designateUpDesc, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(designateUpDesc, prometheus.GaugeValue, 1)
}
