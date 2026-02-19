package placement

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	placementdb "github.com/vexxhost/openstack_database_exporter/internal/db/placement"
)

var (
	placementUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)

	resourceTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "resource_total"),
		"resource_total",
		[]string{
			"hostname",
			"resourcetype",
		},
		nil,
	)

	resourceAllocationRatioDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "resource_allocation_ratio"),
		"resource_allocation_ratio",
		[]string{
			"hostname",
			"resourcetype",
		},
		nil,
	)

	resourceReservedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "resource_reserved"),
		"resource_reserved",
		[]string{
			"hostname",
			"resourcetype",
		},
		nil,
	)

	resourceUsageDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "resource_usage"),
		"resource_usage",
		[]string{
			"hostname",
			"resourcetype",
		},
		nil,
	)
)

type ResourcesCollector struct {
	db      *sql.DB
	queries *placementdb.Queries
	logger  *slog.Logger
}

func NewResourcesCollector(db *sql.DB, logger *slog.Logger) *ResourcesCollector {
	return &ResourcesCollector{
		db:      db,
		queries: placementdb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "resources",
		),
	}
}

func (c *ResourcesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- placementUpDesc
	ch <- resourceTotalDesc
	ch <- resourceAllocationRatioDesc
	ch <- resourceReservedDesc
	ch <- resourceUsageDesc
}

func (c *ResourcesCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	resources, err := c.queries.GetResourceMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to collect placement resources", "error", err)
		ch <- prometheus.MustNewConstMetric(placementUpDesc, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(placementUpDesc, prometheus.GaugeValue, 1)

	for _, resource := range resources {
		hostname := ""
		if resource.Hostname.Valid {
			hostname = resource.Hostname.String
		}

		resourceType := resource.ResourceType

		allocationRatio := resource.AllocationRatio

		used := resource.Used

		ch <- prometheus.MustNewConstMetric(
			resourceTotalDesc,
			prometheus.GaugeValue,
			float64(resource.Total),
			hostname,
			resourceType,
		)

		ch <- prometheus.MustNewConstMetric(
			resourceAllocationRatioDesc,
			prometheus.GaugeValue,
			allocationRatio,
			hostname,
			resourceType,
		)

		ch <- prometheus.MustNewConstMetric(
			resourceReservedDesc,
			prometheus.GaugeValue,
			float64(resource.Reserved),
			hostname,
			resourceType,
		)

		ch <- prometheus.MustNewConstMetric(
			resourceUsageDesc,
			prometheus.GaugeValue,
			float64(used),
			hostname,
			resourceType,
		)
	}
}
