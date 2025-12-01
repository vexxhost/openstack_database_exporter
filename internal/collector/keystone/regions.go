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
	regionsUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)

	regionsCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "regions"),
		"regions",
		nil,
		nil,
	)
)

type RegionsCollector struct {
	db      *sql.DB
	queries *keystonedb.Queries
	logger  *slog.Logger
}

func NewRegionsCollector(db *sql.DB, logger *slog.Logger) *RegionsCollector {
	return &RegionsCollector{
		db:      db,
		queries: keystonedb.New(db),
		logger: logger.With(
			"namespace", collector.Namespace,
			"subsystem", Subsystem,
			"collector", "regions",
		),
	}
}

func (c *RegionsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- regionsUpDesc
	ch <- regionsCountDesc
}

func (c *RegionsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	regions, err := c.queries.GetRegionMetrics(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(regionsUpDesc, prometheus.GaugeValue, 0)
		c.logger.Error("Failed to get region metrics", "error", err)
		return
	}

	// regions count
	ch <- prometheus.MustNewConstMetric(
		regionsCountDesc,
		prometheus.GaugeValue,
		float64(len(regions)),
	)

	ch <- prometheus.MustNewConstMetric(regionsUpDesc, prometheus.GaugeValue, 1)
}
