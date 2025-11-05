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
	groupsCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "groups"),
		"groups",
		nil,
		nil,
	)
)

type GroupsCollector struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewGroupsCollector(db *sql.DB, logger *slog.Logger) *GroupsCollector {
	return &GroupsCollector{
		db:     db,
		logger: logger,
	}
}

func (c *GroupsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- groupsCountDesc
}

func (c *GroupsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	queries := keystonedb.New(c.db)

	c.collectGroupMetrics(ctx, queries, ch)
}

func (c *GroupsCollector) collectGroupMetrics(ctx context.Context, queries *keystonedb.Queries, ch chan<- prometheus.Metric) {
	groups, err := queries.GetGroupMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to get group metrics", "error", err)
		return
	}

	// groups count
	ch <- prometheus.MustNewConstMetric(
		groupsCountDesc,
		prometheus.GaugeValue,
		float64(len(groups)),
	)
}
