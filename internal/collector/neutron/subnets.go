package neutron

import (
	"context"
	"database/sql"
	"encoding/binary"
	"log/slog"
	"math"
	"net/netip"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
	"go4.org/netipx"
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

	subnetsTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "subnets_total"),
		"subnets_total",
		[]string{
			"ip_version",
			"prefix",
			"prefix_length",
			"project_id",
			"subnet_pool_id",
			"subnet_pool_name",
		},
		nil,
	)

	subnetsFreeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "subnets_free"),
		"subnets_free",
		[]string{
			"ip_version",
			"prefix",
			"prefix_length",
			"project_id",
			"subnet_pool_id",
			"subnet_pool_name",
		},
		nil,
	)

	subnetsUsedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "subnets_used"),
		"subnets_used",
		[]string{
			"ip_version",
			"prefix",
			"prefix_length",
			"project_id",
			"subnet_pool_id",
			"subnet_pool_name",
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
	ch <- subnetsTotalDesc
	ch <- subnetsFreeDesc
	ch <- subnetsUsedDesc
}

func (c *SubnetCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	c.collectSubnets(ctx, ch)
	c.collectIPAvailabilities(ctx, ch)
	c.collectSubnetPools(ctx, ch)
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

// subnetpoolWithSubnets associates a subnet pool with the subnets allocated from it.
type subnetpoolWithSubnets struct {
	subnetPool *neutrondb.GetSubnetPoolsRow
	subnets    []netip.Prefix
}

// getPrefixes returns the CIDR prefixes configured for this subnet pool.
func (s *subnetpoolWithSubnets) getPrefixes() []string {
	prefixesStr := cast.ToString(s.subnetPool.Prefixes)
	if prefixesStr == "" {
		return nil
	}
	return strings.Split(prefixesStr, ",")
}

// subnetpoolsWithSubnets groups subnets by their subnet pool.
func subnetpoolsWithSubnets(pools []neutrondb.GetSubnetPoolsRow, subnets []neutrondb.GetSubnetsRow) ([]subnetpoolWithSubnets, error) {
	subnetPrefixes := make(map[string][]netip.Prefix)
	for _, subnet := range subnets {
		if subnet.SubnetpoolID.String != "" {
			subnetPrefix, err := netip.ParsePrefix(subnet.Cidr)
			if err != nil {
				return nil, err
			}
			subnetPrefixes[subnet.SubnetpoolID.String] = append(subnetPrefixes[subnet.SubnetpoolID.String], subnetPrefix)
		}
	}

	result := make([]subnetpoolWithSubnets, len(pools))
	for i, pool := range pools {
		result[i] = subnetpoolWithSubnets{&pools[i], subnetPrefixes[pool.ID]}
	}
	return result, nil
}

// calculateFreeSubnets counts how many CIDRs of the given prefixLength fit in
// poolPrefix after removing subnetsInPool.
func calculateFreeSubnets(poolPrefix *netip.Prefix, subnetsInPool []netip.Prefix, prefixLength int) (float64, error) {
	builder := netipx.IPSetBuilder{}
	builder.AddPrefix(*poolPrefix)

	for _, subnet := range subnetsInPool {
		builder.RemovePrefix(subnet)
	}

	ipset, err := builder.IPSet()
	if err != nil {
		return 0, err
	}

	count := 0.0
	for _, prefix := range ipset.Prefixes() {
		if int(prefix.Bits()) > prefixLength {
			continue
		}
		count += math.Pow(2, float64(prefixLength-int(prefix.Bits())))
	}
	return count, nil
}

// calculateUsedSubnets counts how many of the given subnets overlap with
// ipPrefix and have exactly the specified prefixLength.
func calculateUsedSubnets(subnets []netip.Prefix, ipPrefix netip.Prefix, prefixLength int) float64 {
	result := make(map[int]int)
	for _, subnet := range subnets {
		if !ipPrefix.Overlaps(subnet) {
			continue
		}
		result[int(subnet.Bits())]++
	}
	return float64(result[prefixLength])
}

func (c *SubnetCollector) collectSubnetPools(ctx context.Context, ch chan<- prometheus.Metric) {
	subnets, err := c.queries.GetSubnets(ctx)
	if err != nil {
		c.logger.Error("failed to query subnets for subnet pools", "error", err)
		return
	}

	pools, err := c.queries.GetSubnetPools(ctx)
	if err != nil {
		c.logger.Error("failed to query subnet pools", "error", err)
		return
	}

	poolsWithSubnets, err := subnetpoolsWithSubnets(pools, subnets)
	if err != nil {
		c.logger.Error("failed to associate subnets with pools", "error", err)
		return
	}

	for _, sp := range poolsWithSubnets {
		prefixes := sp.getPrefixes()
		for _, prefix := range prefixes {
			p, err := netip.ParsePrefix(prefix)
			if err != nil {
				c.logger.Error("failed to parse prefix", "prefix", prefix, "error", err)
				continue
			}

			for prefixLen := sp.subnetPool.MinPrefixlen; prefixLen <= sp.subnetPool.MaxPrefixlen; prefixLen++ {
				if prefixLen < int32(p.Bits()) {
					continue
				}

				labels := []string{
					cast.ToString(sp.subnetPool.IpVersion),
					prefix,
					cast.ToString(prefixLen),
					sp.subnetPool.ProjectID.String,
					sp.subnetPool.ID,
					sp.subnetPool.Name.String,
				}

				totalSubnets := math.Pow(2, float64(prefixLen-int32(p.Bits())))
				ch <- prometheus.MustNewConstMetric(subnetsTotalDesc, prometheus.GaugeValue, totalSubnets, labels...)

				freeSubnets, err := calculateFreeSubnets(&p, sp.subnets, int(prefixLen))
				if err != nil {
					c.logger.Error("failed to calculate free subnets", "error", err)
					continue
				}
				ch <- prometheus.MustNewConstMetric(subnetsFreeDesc, prometheus.GaugeValue, freeSubnets, labels...)

				usedSubnets := calculateUsedSubnets(sp.subnets, p, int(prefixLen))
				ch <- prometheus.MustNewConstMetric(subnetsUsedDesc, prometheus.GaugeValue, usedSubnets, labels...)
			}
		}
	}
}
