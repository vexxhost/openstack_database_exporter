package neutron

import (
	"context"
	"database/sql"
	"log/slog"
	"math"
	"strings"

	"go4.org/netipx"

	"net/netip"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

type subnetpoolWithSubnets struct {
	subnetPool *neutrondb.GetSubnetPoolsRow
	subnets    []netip.Prefix
}

func (s *subnetpoolWithSubnets) getPrefixes() []string {
	prefixes := []string{}
	prefixesStr := cast.ToString(s.subnetPool.Prefixes)
	if prefixesStr != "" {
		prefixes = strings.Split(prefixesStr, ",")
	}
	return prefixes
}

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
		result[i] = subnetpoolWithSubnets{&pool, subnetPrefixes[pool.ID]}
	}
	return result, nil
}

// calculateFreeSubnets : Count how many CIDRs of length prefixLength there are in poolPrefix after removing subnetsInPool
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
	ch <- subnetsTotalDesc
	ch <- subnetsFreeDesc
	ch <- subnetsUsedDesc
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

	sps, err := c.queries.GetSubnetPools(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return err
	}

	sps_with_subnets, err := subnetpoolsWithSubnets(sps, snets)
	if err != nil {
		c.logger.Error("failed to transform subnet pools", "error", err)
		return err
	}

	for _, sp := range sps_with_subnets {
		prefixes := sp.getPrefixes()
		for _, prefix := range prefixes {
			p, err := netip.ParsePrefix(prefix)
			if err != nil {
				c.logger.Error("failed to parse prefix", "error", err)
				return err
			}
			for prefixLen := sp.subnetPool.MinPrefixlen; prefixLen <= sp.subnetPool.MaxPrefixlen; prefixLen++ {
				if prefixLen < int32(p.Bits()) {
					continue
				}
				totalSubnets := math.Pow(2, float64(prefixLen-int32(p.Bits())))
				ch <- prometheus.MustNewConstMetric(
					subnetsTotalDesc,
					prometheus.GaugeValue,
					totalSubnets,
					cast.ToString(sp.subnetPool.IpVersion),
					prefix,
					cast.ToString(prefixLen),
					sp.subnetPool.ProjectID.String,
					sp.subnetPool.ID,
					sp.subnetPool.Name.String,
				)

				freeSubnets, err := calculateFreeSubnets(&p, sp.subnets, int(prefixLen))
				if err != nil {
					c.logger.Error("failed to parse prefix", "error", err)
					return err
				}
				ch <- prometheus.MustNewConstMetric(
					subnetsFreeDesc,
					prometheus.GaugeValue,
					freeSubnets,
					cast.ToString(sp.subnetPool.IpVersion),
					prefix,
					cast.ToString(prefixLen),
					sp.subnetPool.ProjectID.String,
					sp.subnetPool.ID,
					sp.subnetPool.Name.String,
				)

				usedSubnets := calculateUsedSubnets(sp.subnets, p, int(prefixLen))
				ch <- prometheus.MustNewConstMetric(
					subnetsUsedDesc,
					prometheus.GaugeValue,
					usedSubnets,
					cast.ToString(sp.subnetPool.IpVersion),
					prefix,
					cast.ToString(prefixLen),
					sp.subnetPool.ProjectID.String,
					sp.subnetPool.ID,
					sp.subnetPool.Name.String,
				)
			}
		}
	}

	return nil
}
