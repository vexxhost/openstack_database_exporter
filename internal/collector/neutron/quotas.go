package neutron

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/project"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

// Neutron default quota values per resource type.
var neutronDefaultQuotas = map[string]int32{
	"floatingip":          50,
	"network":             100,
	"port":                500,
	"rbac_policy":         10,
	"router":              10,
	"security_group":      10,
	"security_group_rule": 100,
	"subnet":              100,
	"subnetpool":          -1,
}

// quotaResources is the ordered list of resource types to emit.
var quotaResources = []string{
	"floatingip",
	"network",
	"port",
	"rbac_policy",
	"router",
	"security_group",
	"security_group_rule",
	"subnet",
	"subnetpool",
}

var quotaDescs map[string]*prometheus.Desc

func init() {
	quotaDescs = make(map[string]*prometheus.Desc, len(quotaResources))
	for _, r := range quotaResources {
		quotaDescs[r] = prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, Subsystem, "quota_"+r),
			"quota_"+r,
			[]string{"tenant", "tenant_id", "type"},
			nil,
		)
	}
}

type QuotaCollector struct {
	db              *sql.DB
	queries         *neutrondb.Queries
	logger          *slog.Logger
	projectResolver *project.Resolver
}

func NewQuotaCollector(db *sql.DB, logger *slog.Logger, projectResolver *project.Resolver) *QuotaCollector {
	return &QuotaCollector{
		db:      db,
		queries: neutrondb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "quotas",
		),
		projectResolver: projectResolver,
	}
}

func (c *QuotaCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range quotaDescs {
		ch <- desc
	}
}

func (c *QuotaCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Get explicit quota limits from DB
	quotaLimits, err := c.queries.GetQuotas(ctx)
	if err != nil {
		c.logger.Error("failed to query quotas", "error", err)
		return
	}

	// Get resource counts per project
	resourceCounts, err := c.queries.GetResourceCountsByProject(ctx)
	if err != nil {
		c.logger.Error("failed to query resource counts", "error", err)
		return
	}

	// Build per-project quota limit map: project_id -> resource -> limit
	projectLimits := make(map[string]map[string]int32)
	for _, q := range quotaLimits {
		pid := q.ProjectID.String
		if _, ok := projectLimits[pid]; !ok {
			projectLimits[pid] = make(map[string]int32)
		}
		projectLimits[pid][q.Resource.String] = q.Limit.Int32
	}

	// Build per-project resource usage map: project_id -> resource -> count
	projectUsage := make(map[string]map[string]int64)
	for _, rc := range resourceCounts {
		pid := rc.ProjectID.String
		if _, ok := projectUsage[pid]; !ok {
			projectUsage[pid] = make(map[string]int64)
		}
		projectUsage[pid][rc.Resource] = rc.Cnt
	}

	// Emit metrics for all active Keystone projects, using DB overrides where available.
	// This filters out stale/deleted projects that remain in the quota DB.
	keystoneProjects := c.projectResolver.AllProjects()
	keystoneAvailable := len(keystoneProjects) > 0

	// Build the set of project IDs to emit: all Keystone projects (if available),
	// otherwise fall back to DB projects only.
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
		for pid := range projectLimits {
			name, _ := c.projectResolver.Resolve(pid)
			projectsToEmit = append(projectsToEmit, projectEntry{id: pid, name: name})
		}
	}

	for _, p := range projectsToEmit {
		limits := projectLimits[p.id]
		usage := projectUsage[p.id]

		for _, resource := range quotaResources {
			desc := quotaDescs[resource]

			// Limit: explicit override or default
			limit := neutronDefaultQuotas[resource]
			if limits != nil {
				if explicitLimit, ok := limits[resource]; ok {
					limit = explicitLimit
				}
			}

			// Used: resource count or 0
			var used int64
			if usage != nil {
				used = usage[resource]
			}

			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(limit), p.name, p.id, "limit")
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 0, p.name, p.id, "reserved")
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(used), p.name, p.id, "used")
		}
	}
}
