package keystone

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
	keystonedb "github.com/vexxhost/openstack_database_exporter/internal/db/keystone"
)

var (
	regionsCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "regions"),
		"regions",
		nil,
		nil,
	)
)

type RegionsCollector struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRegionsCollector(db *sql.DB, logger *slog.Logger) *RegionsCollector {
	return &RegionsCollector{
		db:     db,
		logger: logger,
	}
}

func (c *RegionsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- regionsCountDesc
}

func (c *RegionsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	queries := keystonedb.New(c.db)

	c.collectRegionMetrics(ctx, queries, ch)
}

func (c *RegionsCollector) collectRegionMetrics(ctx context.Context, queries *keystonedb.Queries, ch chan<- prometheus.Metric) {
	regions, err := queries.GetRegionMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to get region metrics", "error", err)
		return
	}

	// regions count
	ch <- prometheus.MustNewConstMetric(
		regionsCountDesc,
		prometheus.GaugeValue,
		float64(len(regions)),
	)
}