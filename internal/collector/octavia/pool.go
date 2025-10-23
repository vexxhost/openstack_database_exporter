package octavia

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
	octaviadb "github.com/vexxhost/openstack_database_exporter/internal/db/octavia"
)

var (
	poolStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "pool_status"),
		"pool_status",
		[]string{
			"id",
			"provisioning_status",
			"name",
			"loadbalancers",
			"protocol",
			"lb_algorithm",
			"operating_status",
			"project_id",
		},
		nil,
	)

	totalPoolsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "total_pools"),
		"total_pools",
		nil,
		nil,
	)
)

type PoolCollector struct {
	db      *sql.DB
	queries *octaviadb.Queries
	logger  *slog.Logger
}

func NewPoolCollector(db *sql.DB, logger *slog.Logger) *PoolCollector {
	return &PoolCollector{
		db:      db,
		queries: octaviadb.New(db),
		logger: logger.With(
			"namespace", collector.Namespace,
			"subsystem", Subsystem,
			"collector", "pool",
		),
	}
}

func (c *PoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- poolStatusDesc
	ch <- totalPoolsDesc
}

func (c *PoolCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	pools, err := c.queries.GetAllPools(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return
	}

	for _, pool := range pools {
		ch <- prometheus.MustNewConstMetric(
			poolStatusDesc,
			prometheus.GaugeValue,
			collector.StatusToValue(pool.ProvisioningStatus, []string{
				"ACTIVE",
				"DELETED",
				"ERROR",
				"PENDING_CREATE",
				"PENDING_UPDATE",
				"PENDING_DELETE",
			}),
			pool.ID,
			pool.ProvisioningStatus,
			pool.Name.String,
			pool.LoadBalancerID.String,
			pool.Protocol,
			pool.LbAlgorithm,
			pool.OperatingStatus,
			pool.ProjectID.String,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		totalPoolsDesc,
		prometheus.GaugeValue,
		float64(len(pools)),
	)
}
