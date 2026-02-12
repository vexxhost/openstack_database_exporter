package nova

import (
	"context"
	"log/slog"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	"github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
)

// ComputeNodesCollector collects metrics about Nova compute nodes
type ComputeNodesCollector struct {
	logger             *slog.Logger
	novaDB             *nova.Queries
	novaAPIDB          *nova_api.Queries
	computeNodeMetrics map[string]*prometheus.Desc
}

// NewComputeNodesCollector creates a new compute nodes collector
func NewComputeNodesCollector(logger *slog.Logger, novaDB *nova.Queries, novaAPIDB *nova_api.Queries) *ComputeNodesCollector {
	return &ComputeNodesCollector{
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "compute_nodes",
		),
		novaDB:    novaDB,
		novaAPIDB: novaAPIDB,
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
	computeNodes, err := c.novaDB.GetComputeNodes(context.Background())
	if err != nil {
		return err
	}

	// Get aggregates info for compute nodes
	aggregates, err := c.novaAPIDB.GetAggregateHosts(context.Background())
	if err != nil {
		c.logger.Error("Failed to get aggregate hosts", "error", err)
	}

	// Build a map of hostname -> aggregates
	hostAggregates := make(map[string][]string)
	for _, agg := range aggregates {
		hostname := agg.Host.String
		if hostname != "" {
			hostAggregates[hostname] = append(hostAggregates[hostname], agg.AggregateName.String)
		}
	}

	for _, node := range computeNodes {
		hostname := node.HypervisorHostname.String
		if hostname == "" {
			continue
		}

		// Get aggregates for this host
		var aggregatesStr string
		if aggList, exists := hostAggregates[hostname]; exists {
			aggregatesStr = strings.Join(aggList, ",")
		}

		availabilityZone := "" // Compute nodes don't have direct AZ assignment

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

		ch <- prometheus.MustNewConstMetric(
			c.computeNodeMetrics["vcpus_available"],
			prometheus.GaugeValue,
			float64(node.Vcpus-node.VcpusUsed),
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
