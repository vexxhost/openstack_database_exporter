package nova

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/project"
	"github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	"github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
	"github.com/vexxhost/openstack_database_exporter/internal/db/placement"
)

// LimitsCollector collects Nova limits metrics using placement as the
// authoritative source for quota usage (covers both DbQuotaDriver and
// UnifiedLimitsDriver).
type LimitsCollector struct {
	logger          *slog.Logger
	novaDB          *nova.Queries
	novaAPIDB       *nova_api.Queries
	placementDB     *placement.Queries
	projectResolver *project.Resolver
	limitsMetrics   map[string]*prometheus.Desc
}

// NewLimitsCollector creates a new limits collector
func NewLimitsCollector(logger *slog.Logger, novaDB *nova.Queries, novaAPIDB *nova_api.Queries, placementDB *placement.Queries, projectResolver *project.Resolver) *LimitsCollector {
	return &LimitsCollector{
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "limits",
		),
		novaDB:          novaDB,
		novaAPIDB:       novaAPIDB,
		placementDB:     placementDB,
		projectResolver: projectResolver,
		limitsMetrics: map[string]*prometheus.Desc{
			"limits_instances_max": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "limits_instances_max"),
				"limits_instances_max",
				[]string{"domain_id", "tenant", "tenant_id"},
				nil,
			),
			"limits_instances_used": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "limits_instances_used"),
				"limits_instances_used",
				[]string{"domain_id", "tenant", "tenant_id"},
				nil,
			),
			"limits_memory_max": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "limits_memory_max"),
				"limits_memory_max",
				[]string{"domain_id", "tenant", "tenant_id"},
				nil,
			),
			"limits_memory_used": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "limits_memory_used"),
				"limits_memory_used",
				[]string{"domain_id", "tenant", "tenant_id"},
				nil,
			),
			"limits_vcpus_max": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "limits_vcpus_max"),
				"limits_vcpus_max",
				[]string{"domain_id", "tenant", "tenant_id"},
				nil,
			),
			"limits_vcpus_used": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "limits_vcpus_used"),
				"limits_vcpus_used",
				[]string{"domain_id", "tenant", "tenant_id"},
				nil,
			),
		},
	}
}

// Describe implements the prometheus.Collector interface
func (c *LimitsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.limitsMetrics {
		ch <- desc
	}
}

// Collect implements the prometheus.Collector interface
func (c *LimitsCollector) Collect(ch chan<- prometheus.Metric) error {
	return c.collectLimitsMetrics(ch)
}

func (c *LimitsCollector) collectLimitsMetrics(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	// Get quotas (limits) from Nova API DB
	quotas, err := c.novaAPIDB.GetQuotas(ctx)
	if err != nil {
		return err
	}

	// Build limits maps by project and resource
	limitsByProject := make(map[string]map[string]float64)
	projectHasQuota := make(map[string]map[string]bool)
	for _, quota := range quotas {
		projectID := quota.ProjectID.String
		resource := quota.Resource
		hardLimit := float64(quota.HardLimit.Int32)

		if limitsByProject[projectID] == nil {
			limitsByProject[projectID] = make(map[string]float64)
			projectHasQuota[projectID] = make(map[string]bool)
		}
		limitsByProject[projectID][resource] = hardLimit
		projectHasQuota[projectID][resource] = true
	}

	// Get usage from placement (authoritative source for quota usage)
	vcpusUsedByProject := make(map[string]float64)
	memoryUsedByProject := make(map[string]float64)
	instanceCountByProject := make(map[string]float64)

	if c.placementDB != nil {
		// Get resource allocations (VCPU, MEMORY_MB) from placement
		allocations, err := c.placementDB.GetAllocationsByProject(ctx)
		if err != nil {
			c.logger.Error("Failed to get allocations from placement", "error", err)
		} else {
			for _, alloc := range allocations {
				used := float64(alloc.Used)
				switch alloc.ResourceType.String {
				case "VCPU":
					vcpusUsedByProject[alloc.ProjectID] = used
				case "MEMORY_MB":
					memoryUsedByProject[alloc.ProjectID] = used
				}
			}
		}

		// Get instance count (consumer count) from placement
		consumerCounts, err := c.placementDB.GetConsumerCountByProject(ctx)
		if err != nil {
			c.logger.Error("Failed to get consumer count from placement", "error", err)
		} else {
			for _, cc := range consumerCounts {
				instanceCountByProject[cc.ProjectID] = float64(cc.InstanceCount)
			}
		}
	} else {
		c.logger.Warn("Placement database not configured, limits_*_used metrics will be 0")
	}

	// Collect all project IDs from both sources
	allProjects := make(map[string]bool)
	for projectID := range limitsByProject {
		allProjects[projectID] = true
	}
	for projectID := range instanceCountByProject {
		allProjects[projectID] = true
	}
	for projectID := range vcpusUsedByProject {
		allProjects[projectID] = true
	}
	for projectID := range memoryUsedByProject {
		allProjects[projectID] = true
	}

	for projectID := range allProjects {
		tenantName, domainID := c.projectResolver.Resolve(projectID)

		// Instances
		instancesMax := float64(10) // Default
		if projectHasQuota[projectID] != nil && projectHasQuota[projectID]["instances"] {
			instancesMax = limitsByProject[projectID]["instances"]
		}

		ch <- prometheus.MustNewConstMetric(
			c.limitsMetrics["limits_instances_max"],
			prometheus.GaugeValue,
			instancesMax,
			domainID, tenantName, projectID,
		)
		ch <- prometheus.MustNewConstMetric(
			c.limitsMetrics["limits_instances_used"],
			prometheus.GaugeValue,
			instanceCountByProject[projectID],
			domainID, tenantName, projectID,
		)

		// Memory
		memoryMax := float64(51200) // Default
		if projectHasQuota[projectID] != nil && projectHasQuota[projectID]["ram"] {
			memoryMax = limitsByProject[projectID]["ram"]
		}

		ch <- prometheus.MustNewConstMetric(
			c.limitsMetrics["limits_memory_max"],
			prometheus.GaugeValue,
			memoryMax,
			domainID, tenantName, projectID,
		)
		ch <- prometheus.MustNewConstMetric(
			c.limitsMetrics["limits_memory_used"],
			prometheus.GaugeValue,
			memoryUsedByProject[projectID],
			domainID, tenantName, projectID,
		)

		// VCPUs
		vcpusMax := float64(20) // Default
		if projectHasQuota[projectID] != nil && projectHasQuota[projectID]["cores"] {
			vcpusMax = limitsByProject[projectID]["cores"]
		}

		ch <- prometheus.MustNewConstMetric(
			c.limitsMetrics["limits_vcpus_max"],
			prometheus.GaugeValue,
			vcpusMax,
			domainID, tenantName, projectID,
		)
		ch <- prometheus.MustNewConstMetric(
			c.limitsMetrics["limits_vcpus_used"],
			prometheus.GaugeValue,
			vcpusUsedByProject[projectID],
			domainID, tenantName, projectID,
		)
	}

	return nil
}
