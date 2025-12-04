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
	db      *sql.DB
	queries *keystonedb.Queries
	logger  *slog.Logger
}

func NewGroupsCollector(db *sql.DB, logger *slog.Logger) *GroupsCollector {
	return &GroupsCollector{
		db:      db,
		queries: keystonedb.New(db),
		logger: logger.With(
			"namespace", collector.Namespace,
			"subsystem", Subsystem,
			"collector", "groups",
		),
	}
}

func (c *GroupsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- groupsCountDesc
}

func (c *GroupsCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	groups, err := c.queries.GetGroupMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to query groups", "error", err)
		return err
	}

	// groups count
	ch <- prometheus.MustNewConstMetric(
		groupsCountDesc,
		prometheus.GaugeValue,
		float64(len(groups)),
	)

	return nil
}
