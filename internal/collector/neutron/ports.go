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
	portDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "port"),
		"port",
		[]string{
			"id",
			"mac_address",
			"device_owner",
			"status",
			"network_id",
			"admin_state_up",
			"binding_vif_type",
			"fixed_ips",
		},
		nil,
	)
	portsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "ports"),
		"ports",
		nil,
		nil,
	)
	lbaasPortsInactiveDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "ports_lb_not_active"),
		"ports_lb_not_active",
		nil,
		nil,
	)
	portsWithNoIPDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "ports_no_ips"),
		"ports_no_ips",
		nil,
		nil,
	)
)

type PortCollector struct {
	db      *sql.DB
	queries *neutrondb.Queries
	logger  *slog.Logger
}

func NewPortCollector(db *sql.DB, logger *slog.Logger) *PortCollector {
	return &PortCollector{
		db:      db,
		queries: neutrondb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "ports",
		),
	}
}

func (c *PortCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- portDesc
	ch <- portsDesc
	ch <- lbaasPortsInactiveDesc
	ch <- portsWithNoIPDesc
}

func (c *PortCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	lbaasPortsInactive := float64(0)
	portsWithNoIP := float64(0)

	ports, err := c.queries.GetPorts(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		portsDesc,
		prometheus.GaugeValue,
		float64(len(ports)),
	)

	for _, port := range ports {
		// Mimicked from: https://github.com/openstack-exporter/openstack-exporter/blob/c298071b2ea4749ad2c9de4184d72308fed7bf2b/exporters/neutron.go#L294
		if port.DeviceOwner == "neutron:LOADBALANCERV2" && port.Status != "ACTIVE" {
			lbaasPortsInactive++
		}
		if port.Status == "ACTIVE" && cast.ToString(port.FixedIps) == "" {
			portsWithNoIP++
		}

		ch <- prometheus.MustNewConstMetric(
			portDesc,
			prometheus.GaugeValue,
			cast.ToFloat64(1),
			port.ID,
			port.MacAddress,
			port.DeviceOwner,
			port.Status,
			port.NetworkID,
			cast.ToString(port.AdminStateUp),
			port.BindingVifType.String,
			cast.ToString(port.FixedIps),
		)
	}

	ch <- prometheus.MustNewConstMetric(
		lbaasPortsInactiveDesc,
		prometheus.GaugeValue,
		lbaasPortsInactive,
	)
	ch <- prometheus.MustNewConstMetric(
		portsWithNoIPDesc,
		prometheus.GaugeValue,
		portsWithNoIP,
	)
	return nil
}
