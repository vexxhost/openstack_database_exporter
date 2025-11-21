package keystone

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
	keystonedb "github.com/vexxhost/openstack_database_exporter/internal/db/keystone"
)

const Subsystem = "identity"

var (
	keystoneUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)

	domainsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "domains"),
		"domains",
		nil,
		nil,
	)

	domainInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "domain_info"),
		"domain_info",
		[]string{
			"description",
			"enabled",
			"id",
			"name",
		},
		nil,
	)

	groupsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "groups"),
		"groups",
		nil,
		nil,
	)

	projectInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "project_info"),
		"project_info",
		[]string{
			"description",
			"domain_id",
			"enabled",
			"id",
			"is_domain",
			"name",
			"parent_id",
			"tags",
		},
		nil,
	)

	projectsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "projects"),
		"projects",
		nil,
		nil,
	)

	regionsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "regions"),
		"regions",
		nil,
		nil,
	)

	usersDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "users"),
		"users",
		nil,
		nil,
	)
)

type IdentityCollector struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewIdentityCollector(db *sql.DB, logger *slog.Logger) *IdentityCollector {
	return &IdentityCollector{
		db:     db,
		logger: logger,
	}
}

func (c *IdentityCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- keystoneUpDesc
	ch <- domainsDesc
	ch <- domainInfoDesc
	ch <- groupsDesc
	ch <- projectInfoDesc
	ch <- projectsDesc
	ch <- regionsDesc
	ch <- usersDesc
}

func (c *IdentityCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	queries := keystonedb.New(c.db)

	// Set up metric = 1 (service is up)
	ch <- prometheus.MustNewConstMetric(
		keystoneUpDesc,
		prometheus.GaugeValue,
		1,
	)

	// Collect domain metrics
	c.collectDomainMetrics(ctx, queries, ch)

	// Collect project metrics
	c.collectProjectMetrics(ctx, queries, ch)

	// Collect group metrics
	c.collectGroupMetrics(ctx, queries, ch)

	// Collect region metrics
	c.collectRegionMetrics(ctx, queries, ch)

	// Collect user metrics
	c.collectUserMetrics(ctx, queries, ch)
}

func (c *IdentityCollector) collectDomainMetrics(ctx context.Context, queries *keystonedb.Queries, ch chan<- prometheus.Metric) {
	domains, err := queries.GetDomainMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to get domain metrics", "error", err)
		return
	}

	// domains count
	ch <- prometheus.MustNewConstMetric(
		domainsDesc,
		prometheus.GaugeValue,
		float64(len(domains)),
	)

	// domain_info metrics
	for _, domain := range domains {
		enabled := "false"
		if domain.Enabled.Valid && domain.Enabled.Bool {
			enabled = "true"
		}

		ch <- prometheus.MustNewConstMetric(
			domainInfoDesc,
			prometheus.GaugeValue,
			1,
			domain.Description,
			enabled,
			domain.ID,
			domain.Name,
		)
	}
}

func (c *IdentityCollector) collectProjectMetrics(ctx context.Context, queries *keystonedb.Queries, ch chan<- prometheus.Metric) {
	projects, err := queries.GetProjectMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to get project metrics", "error", err)
		return
	}

	// projects count
	ch <- prometheus.MustNewConstMetric(
		projectsDesc,
		prometheus.GaugeValue,
		float64(len(projects)),
	)

	// project_info metrics
	for _, project := range projects {
		enabled := "false"
		if project.Enabled.Valid && project.Enabled.Bool {
			enabled = "true"
		}

		isDomain := "false"
		if project.IsDomain {
			isDomain = "true"
		}

		tags := ""
		if project.Tags != nil {
			if tagsStr, ok := project.Tags.(string); ok {
				tags = tagsStr
			}
		}

		ch <- prometheus.MustNewConstMetric(
			projectInfoDesc,
			prometheus.GaugeValue,
			1,
			project.Description,
			project.DomainID,
			enabled,
			project.ID,
			isDomain,
			project.Name,
			project.ParentID,
			tags,
		)
	}
}

func (c *IdentityCollector) collectGroupMetrics(ctx context.Context, queries *keystonedb.Queries, ch chan<- prometheus.Metric) {
	groups, err := queries.GetGroupMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to get group metrics", "error", err)
		return
	}

	// groups count
	ch <- prometheus.MustNewConstMetric(
		groupsDesc,
		prometheus.GaugeValue,
		float64(len(groups)),
	)
}

func (c *IdentityCollector) collectRegionMetrics(ctx context.Context, queries *keystonedb.Queries, ch chan<- prometheus.Metric) {
	regions, err := queries.GetRegionMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to get region metrics", "error", err)
		return
	}

	// regions count
	ch <- prometheus.MustNewConstMetric(
		regionsDesc,
		prometheus.GaugeValue,
		float64(len(regions)),
	)
}

func (c *IdentityCollector) collectUserMetrics(ctx context.Context, queries *keystonedb.Queries, ch chan<- prometheus.Metric) {
	users, err := queries.GetUserMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to get user metrics", "error", err)
		return
	}

	// users count
	ch <- prometheus.MustNewConstMetric(
		usersDesc,
		prometheus.GaugeValue,
		float64(len(users)),
	)
}
