package keystone

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
	keystonedb "github.com/vexxhost/openstack_database_exporter/internal/db/keystone"
)

var (
	domainsCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "domains"),
		"domains",
		nil,
		nil,
	)

	domainsInfoDesc = prometheus.NewDesc(
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
)

type DomainsCollector struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewDomainsCollector(db *sql.DB, logger *slog.Logger) *DomainsCollector {
	return &DomainsCollector{
		db:     db,
		logger: logger,
	}
}

func (c *DomainsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- domainsCountDesc
	ch <- domainsInfoDesc
}

func (c *DomainsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	queries := keystonedb.New(c.db)

	c.collectDomainMetrics(ctx, queries, ch)
}

func (c *DomainsCollector) collectDomainMetrics(ctx context.Context, queries *keystonedb.Queries, ch chan<- prometheus.Metric) {
	domains, err := queries.GetDomainMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to get domain metrics", "error", err)
		return
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
}