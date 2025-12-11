package keystone

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	keystonedb "github.com/vexxhost/openstack_database_exporter/internal/db/keystone"
)

var (
	projectsCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "projects"),
		"projects",
		nil,
		nil,
	)

	projectsInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "project_info"),
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
)

type ProjectsCollector struct {
	db      *sql.DB
	queries *keystonedb.Queries
	logger  *slog.Logger
}

func NewProjectsCollector(db *sql.DB, logger *slog.Logger) *ProjectsCollector {
	return &ProjectsCollector{
		db:      db,
		queries: keystonedb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "projects",
		),
	}
}

func (c *ProjectsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- projectsCountDesc
	ch <- projectsInfoDesc
}

func (c *ProjectsCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	projects, err := c.queries.GetProjectMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to query projects", "error", err)
		return err
	}

	// projects count
	ch <- prometheus.MustNewConstMetric(
		projectsCountDesc,
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
			projectsInfoDesc,
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

	return nil
}
