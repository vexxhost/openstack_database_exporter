package neutron

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

var (
	secGroupsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "security_groups"),
		"security_groups",
		nil,
		nil,
	)
)

type SecurityGroupCollector struct {
	db      *sql.DB
	queries *neutrondb.Queries
	logger  *slog.Logger
}

func NewSecurityGroupCollector(db *sql.DB, logger *slog.Logger) *SecurityGroupCollector {
	return &SecurityGroupCollector{
		db:      db,
		queries: neutrondb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "security_groups",
		),
	}
}

func (c *SecurityGroupCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- secGroupsDesc
}

func (c *SecurityGroupCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	sgs, err := c.queries.GetSecurityGroups(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		secGroupsDesc,
		prometheus.GaugeValue,
		float64(len(sgs)),
	)
	return nil
}
