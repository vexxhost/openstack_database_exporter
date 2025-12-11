package placement

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"

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

		// Convert allocation_ratio from string to float64
		allocationRatio, err := strconv.ParseFloat(resource.AllocationRatio, 64)
		if err != nil {
			c.logger.Warn("Failed to parse allocation ratio", "value", resource.AllocationRatio, "error", err)
			allocationRatio = 1.0 // default value
		}

		// Convert used from interface{} to int64 (mysql returns it as []uint8)
		used := int64(0)
		if resource.Used != nil {
			switch v := resource.Used.(type) {
			case int64:
				used = v
			case []uint8:
				// MySQL returns large numbers as []uint8
				usedStr := string(v)
				if parsedUsed, err := strconv.ParseInt(usedStr, 10, 64); err == nil {
					used = parsedUsed
				}
			}
		}

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
