package nova

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	"github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
	"github.com/vexxhost/openstack_database_exporter/internal/db/placement"
)

// ComputeNodesCollector collects metrics about Nova compute nodes
type ComputeNodesCollector struct {
	logger             *slog.Logger
	novaDB             *nova.Queries
	novaAPIDB          *nova_api.Queries
	placementDB        *placement.Queries
	computeNodeMetrics map[string]*prometheus.Desc
}

// NewComputeNodesCollector creates a new compute nodes collector
func NewComputeNodesCollector(logger *slog.Logger, novaDB *nova.Queries, novaAPIDB *nova_api.Queries, placementDB *placement.Queries) *ComputeNodesCollector {
	return &ComputeNodesCollector{
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "compute_nodes",
		),
		novaDB:      novaDB,
		novaAPIDB:   novaAPIDB,
		placementDB: placementDB,
		computeNodeMetrics: map[string]*prometheus.Desc{
			"current_workload": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "current_workload"),
				"current_workload",
				[]string{"aggregates", "availability_zone", "hostname"},
				nil,
			),
			"free_disk_bytes": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "free_disk_bytes"),
				"free_disk_bytes",
				[]string{"aggregates", "availability_zone", "hostname"},
				nil,
			),
			"local_storage_available_bytes": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "local_storage_available_bytes"),
				"local_storage_available_bytes",
				[]string{"aggregates", "availability_zone", "hostname"},
				nil,
			),
			"local_storage_used_bytes": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "local_storage_used_bytes"),
				"local_storage_used_bytes",
				[]string{"aggregates", "availability_zone", "hostname"},
				nil,
			),
			"memory_available_bytes": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "memory_available_bytes"),
				"memory_available_bytes",
				[]string{"aggregates", "availability_zone", "hostname"},
				nil,
			),
			"memory_used_bytes": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "memory_used_bytes"),
				"memory_used_bytes",
				[]string{"aggregates", "availability_zone", "hostname"},
				nil,
			),
			"running_vms": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "running_vms"),
				"running_vms",
				[]string{"aggregates", "availability_zone", "hostname"},
				nil,
			),
			"vcpus_available": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "vcpus_available"),
				"vcpus_available",
				[]string{"aggregates", "availability_zone", "hostname"},
				nil,
			),
			"vcpus_used": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "vcpus_used"),
				"vcpus_used",
				[]string{"aggregates", "availability_zone", "hostname"},
				nil,
			),
		},
	}
}

// Describe implements the prometheus.Collector interface
func (c *ComputeNodesCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.computeNodeMetrics {
		ch <- desc
	}
}

// Collect implements the prometheus.Collector interface
func (c *ComputeNodesCollector) Collect(ch chan<- prometheus.Metric) error {
	return c.collectComputeNodeMetrics(ch)
}

func (c *ComputeNodesCollector) collectComputeNodeMetrics(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	computeNodes, err := c.novaDB.GetComputeNodes(ctx)
	if err != nil {
		return err
	}

	// Get aggregate hosts and metadata to build AZ and aggregate maps
	aggHosts, err := c.novaAPIDB.GetAggregateHostMap(ctx)
	if err != nil {
		c.logger.Error("Failed to get aggregate hosts", "error", err)
	}

	aggMeta, err := c.novaAPIDB.GetAggregateMetadata(ctx)
	if err != nil {
		c.logger.Error("Failed to get aggregate metadata", "error", err)
	}

	// Build hostname -> PCPU inventory total map from placement. Used as a
	// fallback for vcpus_available on dedicated-CPU hosts, where
	// compute_nodes.vcpus is 0 because CPUs are tracked as PCPU inventory.
	pcpuTotalByHost := make(map[string]int32)
	if c.placementDB != nil {
		resources, err := c.placementDB.GetResourceMetrics(ctx)
		if err != nil {
			c.logger.Error("Failed to get placement resource metrics", "error", err)
		} else {
			for _, r := range resources {
				if r.ResourceType == "PCPU" && r.Hostname.Valid {
					pcpuTotalByHost[r.Hostname.String] = r.Total
				}
			}
		}
	}

	// Build map: aggregate_id -> metadata keys
	aggMetaKeys := make(map[int32][]string)
	aggAZ := make(map[int32]string)
	for _, m := range aggMeta {
		aggMetaKeys[m.AggregateID] = append(aggMetaKeys[m.AggregateID], m.Key)
		if m.Key == "availability_zone" && m.Value.Valid && m.Value.String != "" {
			aggAZ[m.AggregateID] = m.Value.String
		}
	}

	// isAzAggregate: aggregate has exactly one metadata key and it's "availability_zone"
	isAzAggregate := func(aggID int32) bool {
		keys := aggMetaKeys[aggID]
		if len(keys) == 1 && keys[0] == "availability_zone" {
			return true
		}
		return false
	}

	// Build host -> AZ map and host -> non-AZ aggregate names map
	hostToAZ := make(map[string]string)
	hostToAggregates := make(map[string][]string)
	for _, ah := range aggHosts {
		host := ah.Host.String
		if host == "" {
			continue
		}
		if az, ok := aggAZ[ah.AggregateID]; ok {
			hostToAZ[host] = az
		}
		if !isAzAggregate(ah.AggregateID) {
			name := ah.AggregateName.String
			if name != "" {
				hostToAggregates[host] = append(hostToAggregates[host], name)
			}
		}
	}

	for _, node := range computeNodes {
		hostname := node.HypervisorHostname.String
		if hostname == "" {
			continue
		}

		// Use node.Host (service host) to look up AZ and aggregates
		host := node.Host.String

		availabilityZone := ""
		if az, ok := hostToAZ[host]; ok {
			availabilityZone = az
		}

		var aggregatesStr string
		if aggList, exists := hostToAggregates[host]; exists {
			slices.Sort(aggList)
			aggregatesStr = strings.Join(aggList, ",")
		}

		ch <- prometheus.MustNewConstMetric(
			c.computeNodeMetrics["current_workload"],
			prometheus.GaugeValue,
			float64(node.CurrentWorkload.Int32),
			aggregatesStr, availabilityZone, hostname,
		)

		ch <- prometheus.MustNewConstMetric(
			c.computeNodeMetrics["free_disk_bytes"],
			prometheus.GaugeValue,
			float64(node.FreeDiskGb.Int32)*1024*1024*1024,
			aggregatesStr, availabilityZone, hostname,
		)

		ch <- prometheus.MustNewConstMetric(
			c.computeNodeMetrics["local_storage_available_bytes"],
			prometheus.GaugeValue,
			float64(node.LocalGb-node.LocalGbUsed)*1024*1024*1024,
			aggregatesStr, availabilityZone, hostname,
		)

		ch <- prometheus.MustNewConstMetric(
			c.computeNodeMetrics["local_storage_used_bytes"],
			prometheus.GaugeValue,
			float64(node.LocalGbUsed)*1024*1024*1024,
			aggregatesStr, availabilityZone, hostname,
		)

		ch <- prometheus.MustNewConstMetric(
			c.computeNodeMetrics["memory_available_bytes"],
			prometheus.GaugeValue,
			float64(node.MemoryMb-node.MemoryMbUsed)*1024*1024,
			aggregatesStr, availabilityZone, hostname,
		)

		ch <- prometheus.MustNewConstMetric(
			c.computeNodeMetrics["memory_used_bytes"],
			prometheus.GaugeValue,
			float64(node.MemoryMbUsed)*1024*1024,
			aggregatesStr, availabilityZone, hostname,
		)

		ch <- prometheus.MustNewConstMetric(
			c.computeNodeMetrics["running_vms"],
			prometheus.GaugeValue,
			float64(node.RunningVms.Int32),
			aggregatesStr, availabilityZone, hostname,
		)

		vcpusAvailable := node.Vcpus
		if vcpusAvailable == 0 {
			if pcpu, ok := pcpuTotalByHost[hostname]; ok {
				vcpusAvailable = pcpu
			}
		}
		ch <- prometheus.MustNewConstMetric(
			c.computeNodeMetrics["vcpus_available"],
			prometheus.GaugeValue,
			float64(vcpusAvailable),
			aggregatesStr, availabilityZone, hostname,
		)

		ch <- prometheus.MustNewConstMetric(
			c.computeNodeMetrics["vcpus_used"],
			prometheus.GaugeValue,
			float64(node.VcpusUsed),
			aggregatesStr, availabilityZone, hostname,
		)
	}

	return nil
}
