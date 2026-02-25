package neutron

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

var (
	portDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "port"),
		"port",
		[]string{
			"admin_state_up",
			"binding_vif_type",
			"device_owner",
			"fixed_ips",
			"mac_address",
			"network_id",
			"status",
			"uuid",
		},
		nil,
	)

	portsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "ports"),
		"ports",
		nil,
		nil,
	)

	portsLBNotActiveDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "ports_lb_not_active"),
		"ports_lb_not_active",
		nil,
		nil,
	)

	portsNoIPsDesc = prometheus.NewDesc(
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
	ch <- portsLBNotActiveDesc
	ch <- portsNoIPsDesc
}

func (c *PortCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	ports, err := c.queries.GetPorts(ctx)
	if err != nil {
		c.logger.Error("failed to query ports", "error", err)
		return
	}

	lbNotActive := 0
	noIPs := 0
	for _, p := range ports {
		fixedIPs := dbString(p.FixedIps)

		ch <- prometheus.MustNewConstMetric(
			portDesc,
			prometheus.GaugeValue,
			1,
			strconv.FormatBool(p.AdminStateUp),
			p.BindingVifType.String,
			p.DeviceOwner,
			fixedIPs,
			p.MacAddress,
			p.NetworkID,
			p.Status,
			p.ID,
		)

		if p.DeviceOwner == "neutron:LOADBALANCERV2" && p.Status != "ACTIVE" {
			lbNotActive++
		}
		if fixedIPs == "" && p.IpAllocation.String != "none" {
			noIPs++
		}
	}

	ch <- prometheus.MustNewConstMetric(portsDesc, prometheus.GaugeValue, float64(len(ports)))
	ch <- prometheus.MustNewConstMetric(portsLBNotActiveDesc, prometheus.GaugeValue, float64(lbNotActive))
	ch <- prometheus.MustNewConstMetric(portsNoIPsDesc, prometheus.GaugeValue, float64(noIPs))
}
