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
	db      *sql.DB
	queries *keystonedb.Queries
	logger  *slog.Logger
}

func NewUsersCollector(db *sql.DB, logger *slog.Logger) *UsersCollector {
	return &UsersCollector{
		db:      db,
		queries: keystonedb.New(db),
		logger: logger.With(
			"namespace", collector.Namespace,
			"subsystem", Subsystem,
			"collector", "users",
		),
	}
}

func (c *UsersCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- usersCountDesc
}

func (c *UsersCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	users, err := c.queries.GetUserMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to query users", "error", err)
		return err
	}

	// users count
	ch <- prometheus.MustNewConstMetric(
		usersCountDesc,
		prometheus.GaugeValue,
		float64(len(users)),
	)

	return nil
}
