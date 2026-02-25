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

// QuotasCollector collects metrics about Nova quotas
type QuotasCollector struct {
	logger          *slog.Logger
	novaDB          *nova.Queries
	novaAPIDB       *nova_api.Queries
	placementDB     *placement.Queries
	projectResolver *project.Resolver
	quotaMetrics    map[string]*prometheus.Desc
}

// NewQuotasCollector creates a new quotas collector
func NewQuotasCollector(logger *slog.Logger, novaDB *nova.Queries, novaAPIDB *nova_api.Queries, placementDB *placement.Queries, projectResolver *project.Resolver) *QuotasCollector {
	return &QuotasCollector{
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "quotas",
		),
		novaDB:          novaDB,
		novaAPIDB:       novaAPIDB,
		placementDB:     placementDB,
		projectResolver: projectResolver,
		quotaMetrics: map[string]*prometheus.Desc{
			"quota_cores": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_cores"),
				"quota_cores",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
			"quota_fixed_ips": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_fixed_ips"),
				"quota_fixed_ips",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
			"quota_floating_ips": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_floating_ips"),
				"quota_floating_ips",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
			"quota_injected_file_content_bytes": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_injected_file_content_bytes"),
				"quota_injected_file_content_bytes",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
			"quota_injected_file_path_bytes": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_injected_file_path_bytes"),
				"quota_injected_file_path_bytes",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
			"quota_injected_files": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_injected_files"),
				"quota_injected_files",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
			"quota_instances": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_instances"),
				"quota_instances",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
			"quota_key_pairs": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_key_pairs"),
				"quota_key_pairs",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
			"quota_metadata_items": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_metadata_items"),
				"quota_metadata_items",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
			"quota_ram": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_ram"),
				"quota_ram",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
			"quota_security_group_rules": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_security_group_rules"),
				"quota_security_group_rules",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
			"quota_security_groups": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_security_groups"),
				"quota_security_groups",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
			"quota_server_group_members": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_server_group_members"),
				"quota_server_group_members",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
			"quota_server_groups": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_server_groups"),
				"quota_server_groups",
				[]string{"domain_id", "tenant", "type"},
				nil,
			),
		},
	}
}

// Describe implements the prometheus.Collector interface
func (c *QuotasCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.quotaMetrics {
		ch <- desc
	}
}

// Collect implements the prometheus.Collector interface
func (c *QuotasCollector) Collect(ch chan<- prometheus.Metric) error {
	return c.collectQuotaMetrics(ch)
}

func (c *QuotasCollector) collectQuotaMetrics(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	// Get quotas (hard limits)
	quotas, err := c.novaAPIDB.GetQuotas(ctx)
	if err != nil {
		return err
	}

	// Get usage from placement (authoritative source, quota_usages table is often empty)
	vcpusUsedByProject := make(map[string]float64)
	memoryUsedByProject := make(map[string]float64)
	instanceCountByProject := make(map[string]float64)
	// DISK_GB usage from placement
	diskUsedByProject := make(map[string]float64)

	if c.placementDB != nil {
		allocations, err := c.placementDB.GetAllocationsByProject(ctx)
		if err != nil {
			c.logger.Error("Failed to get allocations from placement for quotas", "error", err)
		} else {
			for _, alloc := range allocations {
				used := float64(alloc.Used)
				switch alloc.ResourceType.String {
				case "VCPU":
					vcpusUsedByProject[alloc.ProjectID] = used
				case "MEMORY_MB":
					memoryUsedByProject[alloc.ProjectID] = used
				case "DISK_GB":
					diskUsedByProject[alloc.ProjectID] = used
				}
			}
		}

		consumerCounts, err := c.placementDB.GetConsumerCountByProject(ctx)
		if err != nil {
			c.logger.Error("Failed to get consumer count from placement for quotas", "error", err)
		} else {
			for _, cc := range consumerCounts {
				instanceCountByProject[cc.ProjectID] = float64(cc.InstanceCount)
			}
		}
	}

	// Build quota limits map
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

	// Get default quota class overrides from DB (class_name = 'default')
	dbDefaults := make(map[string]float64)
	quotaClassDefaults, err := c.novaAPIDB.GetQuotaClassDefaults(ctx)
	if err != nil {
		c.logger.Error("Failed to get quota class defaults", "error", err)
	} else {
		for _, qc := range quotaClassDefaults {
			if qc.Resource.Valid {
				dbDefaults[qc.Resource.String] = float64(qc.HardLimit.Int32)
			}
		}
	}

	// Define default quota values (used when no explicit quota is set)
	// Hardcoded Nova defaults as fallback
	defaultQuotas := map[string]float64{
		"cores":                       20,
		"fixed_ips":                   -1,
		"floating_ips":                -1,
		"injected_file_content_bytes": 10240,
		"injected_file_path_bytes":    255,
		"injected_files":              5,
		"instances":                   10,
		"key_pairs":                   100,
		"metadata_items":              128,
		"ram":                         51200,
		"security_group_rules":        -1,
		"security_groups":             10,
		"server_group_members":        10,
		"server_groups":               10,
	}

	// Override hardcoded defaults with DB quota_classes defaults
	for k, v := range dbDefaults {
		defaultQuotas[k] = v
	}

	// Iterate ALL projects from keystone — default quotas apply to every project
	allProjectInfos := c.projectResolver.AllProjects()

	// Emit metrics for each project and quota type
	for projectID, info := range allProjectInfos {
		tenantName := info.Name
		domainID := info.DomainID

		for quotaType, defaultValue := range defaultQuotas {
			// Get limit: use DB value if explicitly set, otherwise use default
			limit := defaultValue
			if projectHasQuota[projectID] != nil && projectHasQuota[projectID][quotaType] {
				limit = limitsByProject[projectID][quotaType]
			}

			// Get usage from placement for the resources we can map
			var usage float64
			switch quotaType {
			case "cores":
				usage = vcpusUsedByProject[projectID]
			case "ram":
				usage = memoryUsedByProject[projectID]
			case "instances":
				usage = instanceCountByProject[projectID]
			default:
				// Other quota types don't have placement equivalents
				usage = 0
			}

			// Reserved is always 0 (placement doesn't track reservations this way)
			reserved := float64(0)

			// Emit the three metrics (in_use, limit, reserved) for each quota type
			metricName := "quota_" + quotaType
			if desc, exists := c.quotaMetrics[metricName]; exists {
				ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, usage, domainID, tenantName, "in_use")
				ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, limit, domainID, tenantName, "limit")
				ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, reserved, domainID, tenantName, "reserved")
			}
		}
	}

	return nil
}
