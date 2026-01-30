package neutron

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

var (
	subnetDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "subnet"),
		"subnet",
		[]string{
			"id",
			"cidr",
			"gateway_ip",
			"network_id",
			"project_id",
			"enable_dhcp",
			"dns_nameservers",
		},
		nil,
	)
	subnetsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "subnets"),
		"subnets",
		nil,
		nil,
	)
)

type SubnetCollector struct {
	db      *sql.DB
	queries *neutrondb.Queries
	logger  *slog.Logger
}

func NewSubnetCollector(db *sql.DB, logger *slog.Logger) *SubnetCollector {
	return &SubnetCollector{
		db:      db,
		queries: neutrondb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "subnets",
		),
	}
}

func (c *SubnetCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- subnetDesc
	ch <- subnetsDesc
}

func (c *SubnetCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	snets, err := c.queries.GetSubnets(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		subnetsDesc,
		prometheus.GaugeValue,
		float64(len(snets)),
	)

	for _, snet := range snets {
		ch <- prometheus.MustNewConstMetric(
			subnetDesc,
			prometheus.GaugeValue,
			cast.ToFloat64(1),
			snet.ID,
			snet.Cidr,
			snet.GatewayIp.String,
			snet.NetworkID,
			snet.ProjectID.String,
			cast.ToString(snet.EnableDhcp.Bool),
			cast.ToString(snet.DnsNameservers),
		)
	}
	return nil
}
