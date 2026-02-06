package neutron

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	neutronUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)
)

type NetworkingCollector struct {
	db                                *sql.DB
	logger                            *slog.Logger
	networkCollector                  *NetworkCollector
	floatingIPCollector               *FloatingIPCollector
	routerCollector                   *RouterCollector
	portCollector                     *PortCollector
	securityGroupCollector            *SecurityGroupCollector
	subnetCollector                   *SubnetCollector
	haRouterAgentPortBindingCollector *HARouterAgentPortBindingCollector
	miscCollector                     *MiscCollector
}

func NewNetworkingCollector(db *sql.DB, logger *slog.Logger) *NetworkingCollector {
	return &NetworkingCollector{
		db:                                db,
		logger:                            logger,
		networkCollector:                  NewNetworkCollector(db, logger),
		floatingIPCollector:               NewFloatingIPCollector(db, logger),
		routerCollector:                   NewRouterCollector(db, logger),
		portCollector:                     NewPortCollector(db, logger),
		securityGroupCollector:            NewSecurityGroupCollector(db, logger),
		subnetCollector:                   NewSubnetCollector(db, logger),
		haRouterAgentPortBindingCollector: NewHARouterAgentPortBindingCollector(db, logger),
		miscCollector:                     NewMiscCollector(db, logger),
	}
}

func (c *NetworkingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- neutronUpDesc
	c.networkCollector.Describe(ch)
	c.floatingIPCollector.Describe(ch)
	c.routerCollector.Describe(ch)
	c.portCollector.Describe(ch)
	c.securityGroupCollector.Describe(ch)
	c.subnetCollector.Describe(ch)
	c.haRouterAgentPortBindingCollector.Describe(ch)
	c.miscCollector.Describe(ch)
}

func (c *NetworkingCollector) Collect(ch chan<- prometheus.Metric) {
	// Track if any sub-collector fails
	var hasError bool

	// Collect metrics from all sub-collectors
	if err := c.networkCollector.Collect(ch); err != nil {
		hasError = true
	}
	if err := c.floatingIPCollector.Collect(ch); err != nil {
		hasError = true
	}
	if err := c.routerCollector.Collect(ch); err != nil {
		hasError = true
	}
	if err := c.portCollector.Collect(ch); err != nil {
		hasError = true
	}
	if err := c.securityGroupCollector.Collect(ch); err != nil {
		hasError = true
	}
	if err := c.subnetCollector.Collect(ch); err != nil {
		hasError = true
	}
	if err := c.haRouterAgentPortBindingCollector.Collect(ch); err != nil {
		hasError = true
	}
	if err := c.miscCollector.Collect(ch); err != nil {
		hasError = true
	}

	// Emit single up metric based on overall success/failure
	upValue := float64(1)
	if hasError {
		upValue = 0
	}
	ch <- prometheus.MustNewConstMetric(
		neutronUpDesc,
		prometheus.GaugeValue,
		upValue,
	)
}
