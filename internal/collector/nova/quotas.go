package nova

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/project"
	keystonedb "github.com/vexxhost/openstack_database_exporter/internal/db/keystone"
	"github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	"github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
	"github.com/vexxhost/openstack_database_exporter/internal/db/placement"
)

// QuotasCollector collects metrics about Nova quotas
type QuotasCollector struct {
	logger          *slog.Logger
	novaDB          *nova.Queries
	novaAPIDB       *nova_api.Queries
	keystoneDB      *keystonedb.Queries
	keystoneRegion  string
	placementDB     *placement.Queries
	projectResolver *project.Resolver
	quotaMetrics    map[string]*prometheus.Desc
	defaultQuotas   DefaultQuotas
}

// NewQuotasCollector creates a new quotas collector
func NewQuotasCollector(logger *slog.Logger, novaDB *nova.Queries, novaAPIDB *nova_api.Queries, placementDB *placement.Queries, projectResolver *project.Resolver, defaultQuotas DefaultQuotas, keystoneDB *keystonedb.Queries, keystoneRegion string) *QuotasCollector {
	return &QuotasCollector{
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "quotas",
		),
		novaDB:          novaDB,
		novaAPIDB:       novaAPIDB,
		keystoneDB:      keystoneDB,
		keystoneRegion:  keystoneRegion,
		placementDB:     placementDB,
		projectResolver: projectResolver,
		defaultQuotas:   defaultQuotas,
		quotaMetrics: map[string]*prometheus.Desc{
			"quota_cores": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_cores"),
				"quota_cores",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_pcpus": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_pcpus"),
				"quota_pcpus",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_fixed_ips": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_fixed_ips"),
				"quota_fixed_ips",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_floating_ips": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_floating_ips"),
				"quota_floating_ips",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_injected_file_content_bytes": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_injected_file_content_bytes"),
				"quota_injected_file_content_bytes",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_injected_file_path_bytes": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_injected_file_path_bytes"),
				"quota_injected_file_path_bytes",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_injected_files": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_injected_files"),
				"quota_injected_files",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_instances": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_instances"),
				"quota_instances",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_key_pairs": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_key_pairs"),
				"quota_key_pairs",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_metadata_items": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_metadata_items"),
				"quota_metadata_items",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_ram": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_ram"),
				"quota_ram",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_security_group_rules": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_security_group_rules"),
				"quota_security_group_rules",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_security_groups": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_security_groups"),
				"quota_security_groups",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_server_group_members": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_server_group_members"),
				"quota_server_group_members",
				[]string{"tenant", "tenant_id", "type"},
				nil,
			),
			"quota_server_groups": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "quota_server_groups"),
				"quota_server_groups",
				[]string{"tenant", "tenant_id", "type"},
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

	// Get usage from placement (authoritative source, quota_usages table is often empty)
	vcpusUsedByProject := make(map[string]float64)
	pcpusUsedByProject := make(map[string]float64)
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
				case "PCPU":
					pcpusUsedByProject[alloc.ProjectID] = used
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

	// Define default quota values (used when no explicit quota is set)
	// Configurable defaults for instances/cores/ram, hardcoded Nova upstream
	// defaults for the remaining resources.
	defaultQuotaMap := map[string]float64{
		"cores":                       float64(c.defaultQuotas.Cores),
		"pcpus":                       float64(c.defaultQuotas.PinnedCores),
		"fixed_ips":                   -1,
		"floating_ips":                -1,
		"injected_file_content_bytes": 10240,
		"injected_file_path_bytes":    255,
		"injected_files":              5,
		"instances":                   float64(c.defaultQuotas.Instances),
		"key_pairs":                   100,
		"metadata_items":              128,
		"ram":                         float64(c.defaultQuotas.RAM),
		"security_group_rules":        -1,
		"security_groups":             10,
		"server_group_members":        10,
		"server_groups":               10,
	}

	if c.keystoneDB != nil {
		// Unified Limits: read registered_limits (defaults) and project limits from Keystone
		regLimits, err := c.keystoneDB.GetRegisteredLimits(ctx, keystonedb.GetRegisteredLimitsParams{RegionID: sql.NullString{String: c.keystoneRegion, Valid: true}})
		if err != nil {
			c.logger.Error("Failed to get registered limits from keystone", "error", err)
		} else {
			for _, rl := range regLimits {
				if rl.ResourceName.Valid {
					if name, ok := unifiedLimitToNova(rl.ResourceName.String); ok {
						defaultQuotaMap[name] = float64(rl.DefaultLimit)
					}
				}
			}
		}

		projLimits, err := c.keystoneDB.GetProjectLimits(ctx, keystonedb.GetProjectLimitsParams{RegionID: sql.NullString{String: c.keystoneRegion, Valid: true}})
		if err != nil {
			c.logger.Error("Failed to get project limits from keystone", "error", err)
		} else {
			for _, pl := range projLimits {
				if !pl.ProjectID.Valid || !pl.ResourceName.Valid {
					continue
				}
				name, ok := unifiedLimitToNova(pl.ResourceName.String)
				if !ok {
					continue
				}
				projectID := pl.ProjectID.String
				if limitsByProject[projectID] == nil {
					limitsByProject[projectID] = make(map[string]float64)
					projectHasQuota[projectID] = make(map[string]bool)
				}
				limitsByProject[projectID][name] = float64(pl.ResourceLimit)
				projectHasQuota[projectID][name] = true
			}
		}
	} else {
		// Legacy DB quota driver: read from nova_api.quotas and quota_classes
		quotas, err := c.novaAPIDB.GetQuotas(ctx)
		if err != nil {
			return err
		}
		for _, quota := range quotas {
			if !quota.ProjectID.Valid || !quota.HardLimit.Valid {
				continue
			}
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

		quotaClassDefaults, err := c.novaAPIDB.GetQuotaClassDefaults(ctx)
		if err != nil {
			c.logger.Error("Failed to get quota class defaults", "error", err)
		} else {
			for _, qc := range quotaClassDefaults {
				if qc.Resource.Valid && qc.HardLimit.Valid {
					defaultQuotaMap[qc.Resource.String] = float64(qc.HardLimit.Int32)
				}
			}
		}
	}

	// Emit quotas for all active Keystone projects, using DB overrides where available.
	// Falls back to DB-only projects if Keystone is unavailable.
	keystoneProjects := c.projectResolver.AllProjects()
	keystoneAvailable := len(keystoneProjects) > 0

	type projectEntry struct {
		id   string
		name string
	}
	var projectsToEmit []projectEntry

	if keystoneAvailable {
		for pid, info := range keystoneProjects {
			projectsToEmit = append(projectsToEmit, projectEntry{id: pid, name: info.Name})
		}
	} else {
		for pid := range projectHasQuota {
			name, _ := c.projectResolver.Resolve(pid)
			projectsToEmit = append(projectsToEmit, projectEntry{id: pid, name: name})
		}
	}

	for _, p := range projectsToEmit {
		projectID := p.id
		tenantName := p.name

		for quotaType, defaultValue := range defaultQuotaMap {
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
			case "pcpus":
				usage = pcpusUsedByProject[projectID]
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
				ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, usage, tenantName, projectID, "in_use")
				ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, limit, tenantName, projectID, "limit")
				ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, reserved, tenantName, projectID, "reserved")
			}
		}
	}

	return nil
}
