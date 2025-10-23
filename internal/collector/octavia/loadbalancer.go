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
	loadBalancerStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "loadbalancer_status"),
		"loadbalancer_status",
		[]string{
			"id",
			"name",
			"project_id",
			"operating_status",
			"provisioning_status",
			"provider",
			"vip_address",
		},
		nil,
	)

	totalLoadBalancersDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "total_loadbalancers"),
		"total_loadbalancers",
		nil,
		nil,
	)

	upDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)
)

type LoadBalancerCollector struct {
	db      *sql.DB
	queries *octaviadb.Queries
	logger  *slog.Logger
}

func NewLoadBalancerCollector(db *sql.DB, logger *slog.Logger) *LoadBalancerCollector {
	return &LoadBalancerCollector{
		db:      db,
		queries: octaviadb.New(db),
		logger: logger.With(
			"namespace", collector.Namespace,
			"subsystem", Subsystem,
			"collector", "loadbalancer",
		),
	}
}

func (c *LoadBalancerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- loadBalancerStatusDesc
	ch <- totalLoadBalancersDesc
	ch <- upDesc
}

func (c *LoadBalancerCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	loadBalancers, err := c.queries.GetAllLoadBalancersWithVip(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0)

		c.logger.Error("failed to query", "error", err)
		return
	}

	for _, lb := range loadBalancers {
		ch <- prometheus.MustNewConstMetric(
			loadBalancerStatusDesc,
			prometheus.GaugeValue,
			statusToValue(lb.OperatingStatus, []string{
				"ONLINE",
				"DRAINING",
				"OFFLINE",
				"ERROR",
				"NO_MONITOR",
			}),
			lb.ID,
			lb.Name.String,
			lb.ProjectID.String,
			lb.OperatingStatus,
			lb.ProvisioningStatus,
			lb.Provider.String,
			lb.VipAddress.String,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		totalLoadBalancersDesc,
		prometheus.GaugeValue,
		float64(len(loadBalancers)),
	)

	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 1)
}
