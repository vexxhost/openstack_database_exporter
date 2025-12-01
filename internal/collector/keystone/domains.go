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
	domainsUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)

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
	db      *sql.DB
	queries *keystonedb.Queries
	logger  *slog.Logger
}

func NewDomainsCollector(db *sql.DB, logger *slog.Logger) *DomainsCollector {
	return &DomainsCollector{
		db:      db,
		queries: keystonedb.New(db),
		logger: logger.With(
			"namespace", collector.Namespace,
			"subsystem", Subsystem,
			"collector", "domains",
		),
	}
}

func (c *DomainsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- domainsUpDesc
	ch <- domainsCountDesc
	ch <- domainsInfoDesc
}

func (c *DomainsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	domains, err := c.queries.GetDomainMetrics(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(domainsUpDesc, prometheus.GaugeValue, 0)
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

	ch <- prometheus.MustNewConstMetric(domainsUpDesc, prometheus.GaugeValue, 1)
}
