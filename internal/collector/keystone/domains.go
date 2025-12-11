package keystone

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	keystonedb "github.com/vexxhost/openstack_database_exporter/internal/db/keystone"
)

var (
	domainsCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "domains"),
		"domains",
		nil,
		nil,
	)

	domainsInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "domain_info"),
		"domain_info",
		[]string{
			"description",
			"enabled",
			"id",
			"name",
		},
		nil,
	)
)

type DomainsCollector struct {
	db      *sql.DB
	queries *keystonedb.Queries
	logger  *slog.Logger
}

func NewDomainsCollector(db *sql.DB, logger *slog.Logger) *DomainsCollector {
	return &DomainsCollector{
		db:      db,
		queries: keystonedb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "domains",
		),
	}
}

func (c *DomainsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- domainsCountDesc
	ch <- domainsInfoDesc
}

func (c *DomainsCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	domains, err := c.queries.GetDomainMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to query domains", "error", err)
		return err
	}

	// domains count
	ch <- prometheus.MustNewConstMetric(
		domainsCountDesc,
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
			domainsInfoDesc,
			prometheus.GaugeValue,
			1,
			domain.Description,
			enabled,
			domain.ID,
			domain.Name,
		)
	}

	return nil
}
