package keystone

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	keystonedb "github.com/vexxhost/openstack_database_exporter/internal/db/keystone"
)

var (
	projectTagsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "project_tag"),
		"project_tag",
		[]string{
			"project_id",
			"tag",
		},
		nil,
	)
)

type ProjectTagsCollector struct {
	db      *sql.DB
	queries *keystonedb.Queries
	logger  *slog.Logger
}

func NewProjectTagsCollector(db *sql.DB, logger *slog.Logger) *ProjectTagsCollector {
	return &ProjectTagsCollector{
		db:      db,
		queries: keystonedb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "project_tags",
		),
	}
}

func (c *ProjectTagsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- projectTagsDesc
}

func (c *ProjectTagsCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	tags, err := c.queries.GetProjectTags(ctx)
	if err != nil {
		c.logger.Error("Failed to query project tags", "error", err)
		return err
	}

	for _, tag := range tags {
		ch <- prometheus.MustNewConstMetric(
			projectTagsDesc,
			prometheus.GaugeValue,
			1,
			tag.ProjectID,
			tag.Name,
		)
	}

	return nil
}
