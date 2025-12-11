package keystone

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	keystonedb "github.com/vexxhost/openstack_database_exporter/internal/db/keystone"
)

var (
	regionsCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "regions"),
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
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "regions",
		),
	}
}

func (c *RegionsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- regionsCountDesc
}

func (c *RegionsCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	regions, err := c.queries.GetRegionMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to query regions", "error", err)
		return err
	}

	// regions count
	ch <- prometheus.MustNewConstMetric(
		regionsCountDesc,
		prometheus.GaugeValue,
		float64(len(regions)),
	)

	return nil
}
