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

	floatingsIPsAssociatedNotActive = prometheus.NewDesc(
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
	ch <- floatingsIPsAssociatedNotActive
}

func (c *FloatingIPCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	anaFips := 0

	fips, err := c.queries.GetFloatingIPs(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		floatingIPsDesc,
		prometheus.GaugeValue,
		float64(len(fips)),
	)

	for _, fip := range fips {
		if fip.FixedIpAddress.Valid && fip.Status.String != "ACTIVE" {
			anaFips += 1
		}

		ch <- prometheus.MustNewConstMetric(
			floatingIPDesc,
			prometheus.GaugeValue,
			cast.ToFloat64(1),
			fip.FloatingIpAddress,
			fip.FloatingNetworkID,
			fip.ID,
			fip.ProjectID.String,
			fip.RouterID.String,
			fip.Status.String,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		floatingsIPsAssociatedNotActive,
		prometheus.GaugeValue,
		float64(anaFips),
	)
	return nil
}
