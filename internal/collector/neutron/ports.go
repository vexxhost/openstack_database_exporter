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
}

func (c *PortCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

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
	return nil
}
