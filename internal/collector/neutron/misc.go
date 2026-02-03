package neutron

import (
	"context"
	"database/sql"
	"log/slog"
	"math/big"
	"net"

	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

// ipToBigInt converts an IP address to a big.Int
func ipToBigInt(ip net.IP) (*big.Int, error) {
	if ip == nil {
		return nil, fmt.Errorf("nil IP")
	}

	ip = ip.To16()
	if ip == nil {
		return nil, fmt.Errorf("invalid IP")
	}

	return new(big.Int).SetBytes(ip), nil
}

// CIDRSize returns the total number of IPs in a CIDR
// Equivalent to netaddr.IPNetwork(cidr).size in Python
func CIDRSize(cidr string) (*big.Int, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	ones, bits := ipNet.Mask.Size()

	// bits = 32 for IPv4, 128 for IPv6
	hostBits := bits - ones

	// size = 2^(hostBits)
	size := new(big.Int).Exp(
		big.NewInt(2),
		big.NewInt(int64(hostBits)),
		nil,
	)

	return size, nil
}

// IPRangeSize returns the inclusive size of an IP range
// Equivalent to: netaddr.IPRange(first, last).size in Python
func IPRangeSize(firstIP, lastIP string) (*big.Int, error) {
	first := net.ParseIP(firstIP)
	last := net.ParseIP(lastIP)

	if first == nil || last == nil {
		return nil, fmt.Errorf("invalid IP address")
	}

	firstInt, err := ipToBigInt(first)
	if err != nil {
		return nil, err
	}

	lastInt, err := ipToBigInt(last)
	if err != nil {
		return nil, err
	}

	// last - first + 1 (inclusive range)
	size := new(big.Int).Sub(lastInt, firstInt)
	size.Add(size, big.NewInt(1))

	return size, nil
}

var (
	networkIPAvailabilitiesUsed = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "network_ip_availabilities_used"),
		"network_ip_availabilities_used",
		[]string{
			"network_id",
			"subnet_id",
			"project_id",
			"network_name",
			"subnet_name",
			"cidr",
			"ip_version",
		},
		nil,
	)
	networkIPAvailabilitiesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "network_ip_availabilities_total"),
		"network_ip_availabilities_total",
		[]string{
			"network_id",
			"subnet_id",
			"project_id",
			"network_name",
			"subnet_name",
			"cidr",
			"ip_version",
		},
		nil,
	)
)

type MiscCollector struct {
	db      *sql.DB
	queries *neutrondb.Queries
	logger  *slog.Logger
}

type NetworkAvailabilityTotal struct {
	NetworkID   string
	SubnetID    string
	ProjectID   string
	NetworkName string
	SubnetName  string
	Cidr        string
	IpVersion   string
	Total       float64
}

func NewMiscCollector(db *sql.DB, logger *slog.Logger) *MiscCollector {
	return &MiscCollector{
		db:      db,
		queries: neutrondb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "subnets",
		),
	}
}

func (c *MiscCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- networkIPAvailabilitiesUsed
	ch <- networkIPAvailabilitiesTotal
}

func (c *MiscCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	nipaus, err := c.queries.GetNetworkIPAvailabilitiesUsed(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return err
	}

	for _, nipau := range nipaus {
		ch <- prometheus.MustNewConstMetric(
			networkIPAvailabilitiesUsed,
			prometheus.GaugeValue,
			cast.ToFloat64(nipau.AllocationCount),
			nipau.NetworkID.String,
			nipau.SubnetID,
			nipau.ProjectID.String,
			nipau.NetworkName.String,
			nipau.SubnetName.String,
			nipau.Cidr,
			cast.ToString(nipau.IpVersion),
		)
	}

	nipats, err := c.queries.GetNetworkIPAvailabilitiesTotal(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return err
	}

	seenSubnets := make(map[string]*NetworkAvailabilityTotal, len(nipats))
	for _, nipat := range nipats {
		var size *big.Int
		var err error

		if nipat.LastIp.Valid {
			size, err = IPRangeSize(nipat.FirstIp.String, nipat.LastIp.String)
		} else {
			size, err = CIDRSize(nipat.Cidr)
		}
		if err != nil {
			c.logger.Error("failed to compute range", "error", err)
			return err
		}

		sizeValue, _ := new(big.Float).SetInt(size).Float64()

		if _, exists := seenSubnets[nipat.SubnetID]; !exists {
			seenSubnets[nipat.SubnetID] = &NetworkAvailabilityTotal{
				NetworkID:   nipat.NetworkID,
				SubnetID:    nipat.SubnetID,
				ProjectID:   nipat.ProjectID.String,
				NetworkName: nipat.NetworkName.String,
				SubnetName:  nipat.SubnetName.String,
				Cidr:        nipat.Cidr,
				IpVersion:   cast.ToString(nipat.IpVersion),
				Total:       0,
			}
		}
		seenSubnets[nipat.SubnetID].Total += sizeValue
	}

	for _, nipat := range seenSubnets {
		ch <- prometheus.MustNewConstMetric(
			networkIPAvailabilitiesTotal,
			prometheus.GaugeValue,
			seenSubnets[nipat.SubnetID].Total,
			nipat.NetworkID,
			nipat.SubnetID,
			nipat.ProjectID,
			nipat.NetworkName,
			nipat.SubnetName,
			nipat.Cidr,
			nipat.IpVersion,
		)
	}
	return nil
}
