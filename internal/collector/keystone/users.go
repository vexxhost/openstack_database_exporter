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
	usersCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "users"),
		"users",
		nil,
		nil,
	)
)

type UsersCollector struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUsersCollector(db *sql.DB, logger *slog.Logger) *UsersCollector {
	return &UsersCollector{
		db:     db,
		logger: logger,
	}
}

func (c *UsersCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- usersCountDesc
}

func (c *UsersCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	queries := keystonedb.New(c.db)

	c.collectUserMetrics(ctx, queries, ch)
}

func (c *UsersCollector) collectUserMetrics(ctx context.Context, queries *keystonedb.Queries, ch chan<- prometheus.Metric) {
	users, err := queries.GetUserMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to get user metrics", "error", err)
		return
	}

	// users count
	ch <- prometheus.MustNewConstMetric(
		usersCountDesc,
		prometheus.GaugeValue,
		float64(len(users)),
	)
}
