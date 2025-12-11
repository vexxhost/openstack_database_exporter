package cinder

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	cinderdb "github.com/vexxhost/openstack_database_exporter/internal/db/cinder"
)

var (
	snapshotsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "snapshots"),
		"snapshots",
		nil,
		nil,
	)
)

type SnapshotsCollector struct {
	db      *sql.DB
	queries *cinderdb.Queries
	logger  *slog.Logger
}

func NewSnapshotsCollector(db *sql.DB, logger *slog.Logger) *SnapshotsCollector {
	return &SnapshotsCollector{
		db:      db,
		queries: cinderdb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "snapshots",
		),
	}
}

func (c *SnapshotsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- snapshotsDesc
}

func (c *SnapshotsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	count, err := c.queries.GetSnapshotCount(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		snapshotsDesc,
		prometheus.GaugeValue,
		float64(count),
	)
}
