package neutron

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

var (
	floatingIPDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "floating_ip"),
		"floating_ip",
		[]string{
			"floating_ip_address",
			"floating_network_id",
			"id",
			"project_id",
			"router_id",
			"status",
		},
		nil,
	)

	floatingIPsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "floating_ips"),
		"floating_ips",
		nil,
		nil,
	)

	floatingIPsAssociatedNotActiveDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "floating_ips_associated_not_active"),
		"floating_ips_associated_not_active",
		nil,
		nil,
	)
)

type FloatingIPCollector struct {
	db      *sql.DB
	queries *neutrondb.Queries
	logger  *slog.Logger
}

func NewFloatingIPCollector(db *sql.DB, logger *slog.Logger) *FloatingIPCollector {
	return &FloatingIPCollector{
		db:      db,
		queries: neutrondb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "floating_ips",
		),
	}
}

func (c *FloatingIPCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- floatingIPDesc
	ch <- floatingIPsDesc
	ch <- floatingIPsAssociatedNotActiveDesc
}

func (c *FloatingIPCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	fips, err := c.queries.GetFloatingIPs(ctx)
	if err != nil {
		c.logger.Error("failed to query floating IPs", "error", err)
		return
	}

	associatedNotActive := 0
	for _, fip := range fips {
		ch <- prometheus.MustNewConstMetric(
			floatingIPDesc,
			prometheus.GaugeValue,
			1,
			fip.FloatingIpAddress,
			fip.FloatingNetworkID,
			fip.ID,
			fip.ProjectID.String,
			fip.RouterID.String,
			fip.Status.String,
		)

		if fip.RouterID.Valid && fip.RouterID.String != "" && fip.Status.String != "ACTIVE" {
			associatedNotActive++
		}
	}

	ch <- prometheus.MustNewConstMetric(floatingIPsDesc, prometheus.GaugeValue, float64(len(fips)))
	ch <- prometheus.MustNewConstMetric(floatingIPsAssociatedNotActiveDesc, prometheus.GaugeValue, float64(associatedNotActive))
}
