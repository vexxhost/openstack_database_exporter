package neutron

import (
	"context"
	"database/sql"
	"encoding/binary"
	"log/slog"
	"net/netip"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

var (
	subnetDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "subnet"),
		"subnet",
		[]string{
			"cidr",
			"dns_nameservers",
			"enable_dhcp",
			"gateway_ip",
			"id",
			"name",
			"network_id",
			"tags",
			"tenant_id",
		},
		nil,
	)

	subnetsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "subnets"),
		"subnets",
		nil,
		nil,
	)

	networkIPAvailabilitiesTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "network_ip_availabilities_total"),
		"network_ip_availabilities_total",
		[]string{
			"cidr",
			"ip_version",
			"network_id",
			"network_name",
			"project_id",
			"subnet_name",
		},
		nil,
	)

	networkIPAvailabilitiesUsedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "network_ip_availabilities_used"),
		"network_ip_availabilities_used",
		[]string{
			"cidr",
			"ip_version",
			"network_id",
			"network_name",
			"project_id",
			"subnet_name",
		},
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
	ch <- networkIPAvailabilitiesTotalDesc
	ch <- networkIPAvailabilitiesUsedDesc
}

func (c *SubnetCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	c.collectSubnets(ctx, ch)
	c.collectIPAvailabilities(ctx, ch)
}

func (c *SubnetCollector) collectSubnets(ctx context.Context, ch chan<- prometheus.Metric) {
	subnets, err := c.queries.GetSubnets(ctx)
	if err != nil {
		c.logger.Error("failed to query subnets", "error", err)
		return
	}

	for _, s := range subnets {
		ch <- prometheus.MustNewConstMetric(
			subnetDesc,
			prometheus.GaugeValue,
			1,
			s.Cidr,
			dbString(s.DnsNameservers),
			strconv.FormatBool(s.EnableDhcp.Bool),
			s.GatewayIp.String,
			s.ID,
			s.Name.String,
			s.NetworkID,
			dbString(s.Tags),
			s.ProjectID.String,
		)
	}

	ch <- prometheus.MustNewConstMetric(subnetsDesc, prometheus.GaugeValue, float64(len(subnets)))
}

func (c *SubnetCollector) collectIPAvailabilities(ctx context.Context, ch chan<- prometheus.Metric) {
	// Collect "used" (allocation counts per subnet)
	used, err := c.queries.GetNetworkIPAvailabilitiesUsed(ctx)
	if err != nil {
		c.logger.Error("failed to query IP availability used", "error", err)
		return
	}

	for _, u := range used {
		ch <- prometheus.MustNewConstMetric(
			networkIPAvailabilitiesUsedDesc,
			prometheus.GaugeValue,
			float64(u.AllocationCount),
			u.Cidr,
			strconv.Itoa(int(u.IpVersion)),
			u.NetworkID.String,
			u.NetworkName.String,
			u.ProjectID.String,
			u.SubnetName.String,
		)
	}

	// Collect "total" (sum of allocation pool ranges per subnet)
	total, err := c.queries.GetNetworkIPAvailabilitiesTotal(ctx)
	if err != nil {
		c.logger.Error("failed to query IP availability total", "error", err)
		return
	}

	// Group allocation pool ranges by subnet and sum
	type subnetInfo struct {
		cidr        string
		ipVersion   int32
		networkID   string
		networkName string
		projectID   string
		subnetName  string
		totalIPs    int64
	}

	subnetTotals := make(map[string]*subnetInfo) // keyed by subnet_id
	for _, t := range total {
		si, ok := subnetTotals[t.SubnetID]
		if !ok {
			si = &subnetInfo{
				cidr:        t.Cidr,
				ipVersion:   t.IpVersion,
				networkID:   t.NetworkID,
				networkName: t.NetworkName.String,
				projectID:   t.ProjectID.String,
				subnetName:  t.SubnetName.String,
			}
			subnetTotals[t.SubnetID] = si
		}

		if t.FirstIp.Valid && t.LastIp.Valid {
			si.totalIPs += ipRangeSize(t.FirstIp.String, t.LastIp.String)
		}
	}

	for _, si := range subnetTotals {
		ch <- prometheus.MustNewConstMetric(
			networkIPAvailabilitiesTotalDesc,
			prometheus.GaugeValue,
			float64(si.totalIPs),
			si.cidr,
			strconv.Itoa(int(si.ipVersion)),
			si.networkID,
			si.networkName,
			si.projectID,
			si.subnetName,
		)
	}
}

// ipRangeSize returns the number of IPs in the range [firstIP, lastIP] inclusive.
func ipRangeSize(firstIP, lastIP string) int64 {
	first, err := netip.ParseAddr(firstIP)
	if err != nil {
		return 0
	}
	last, err := netip.ParseAddr(lastIP)
	if err != nil {
		return 0
	}

	if first.Is4() && last.Is4() {
		f := first.As4()
		l := last.As4()
		fInt := binary.BigEndian.Uint32(f[:])
		lInt := binary.BigEndian.Uint32(l[:])
		return int64(lInt-fInt) + 1
	}

	// IPv6: convert 16-byte addresses to uint128 and subtract
	if first.Is6() && last.Is6() {
		f := first.As16()
		l := last.As16()
		fHi := binary.BigEndian.Uint64(f[:8])
		fLo := binary.BigEndian.Uint64(f[8:])
		lHi := binary.BigEndian.Uint64(l[:8])
		lLo := binary.BigEndian.Uint64(l[8:])

		// Simple case: if high 64 bits are the same, just subtract low parts
		if fHi == lHi {
			return int64(lLo-fLo) + 1
		}
	}

	return 0
}
