package manila

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
	maniladb "github.com/vexxhost/openstack_database_exporter/internal/db/manila"
)

const Subsystem = "sharev2"

var (
	// Known share statuses from the original openstack-exporter
	shareStatuses = []string{
		"available", "creating", "deleting", "error", "error_deleting",
		"extending", "inactive", "managing", "migrating", "migration_error",
		"restoring", "reverting", "reverting_error", "reverting_to_snapshot",
		"shrinking", "shrinking_error", "soft_deleting", "unmanaging", "updating",
	}

	manilaUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)

	shareGbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "share_gb"),
		"share_gb",
		[]string{
			"availability_zone",
			"id",
			"name",
			"project_id",
			"share_proto",
			"share_type",
			"share_type_name",
			"status",
		},
		nil,
	)

	shareStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "share_status"),
		"share_status",
		[]string{
			"id",
			"name",
			"project_id",
			"share_proto",
			"share_type",
			"share_type_name",
			"size",
			"status",
		},
		nil,
	)

	shareStatusCounterDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "share_status_counter"),
		"share_status_counter",
		[]string{
			"status",
		},
		nil,
	)

	sharesCounterDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "shares_counter"),
		"shares_counter",
		nil,
		nil,
	)
)

type SharesCollector struct {
	db      *sql.DB
	queries *maniladb.Queries
	logger  *slog.Logger
}

func NewSharesCollector(db *sql.DB, logger *slog.Logger) *SharesCollector {
	return &SharesCollector{
		db:      db,
		queries: maniladb.New(db),
		logger: logger.With(
			"component", "manila_shares_collector",
		),
	}
}

func (c *SharesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- manilaUpDesc
	ch <- shareGbDesc
	ch <- shareStatusDesc
	ch <- shareStatusCounterDesc
	ch <- sharesCounterDesc
}

func (c *SharesCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	shares, err := c.queries.GetShareMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to collect manila shares", "error", err)
		ch <- prometheus.MustNewConstMetric(manilaUpDesc, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(manilaUpDesc, prometheus.GaugeValue, 1)

	// Count shares by status for status counter metrics
	statusCounts := make(map[string]int)
	for _, status := range shareStatuses {
		statusCounts[status] = 0
	}

	totalShares := 0

	for _, share := range shares {
		// Extract values with proper null handling
		shareID := share.ID
		name := ""
		if share.Name.Valid {
			name = share.Name.String
		}
		projectID := ""
		if share.ProjectID.Valid {
			projectID = share.ProjectID.String
		}
		size := int32(0)
		if share.Size.Valid {
			size = share.Size.Int32
		}
		shareProto := ""
		if share.ShareProto.Valid {
			shareProto = share.ShareProto.String
		}
		status := ""
		if share.Status.Valid {
			status = share.Status.String
		}
		shareTypeName := share.ShareTypeName
		availabilityZone := share.AvailabilityZone

		// For share_type label, use availability_zone if available, otherwise empty
		shareType := availabilityZone

		totalShares++

		// Count status for counter metrics
		if status != "" {
			statusCounts[status]++
		}

		// share_gb metric - size in GB per share
		ch <- prometheus.MustNewConstMetric(
			shareGbDesc,
			prometheus.GaugeValue,
			float64(size),
			availabilityZone,
			shareID,
			name,
			projectID,
			shareProto,
			shareType,
			shareTypeName,
			status,
		)

		// share_status metric - status indicator per share
		statusValue := 0.0
		if status != "" {
			statusValue = 1.0
		}

		// Convert size to string properly
		sizeStr := "0"
		if share.Size.Valid {
			sizeStr = fmt.Sprintf("%d", share.Size.Int32)
		}

		ch <- prometheus.MustNewConstMetric(
			shareStatusDesc,
			prometheus.GaugeValue,
			statusValue,
			shareID,
			name,
			projectID,
			shareProto,
			shareType,
			shareTypeName,
			sizeStr,
			status,
		)
	}

	// Emit status counter metrics for all known statuses
	for _, status := range shareStatuses {
		ch <- prometheus.MustNewConstMetric(
			shareStatusCounterDesc,
			prometheus.GaugeValue,
			float64(statusCounts[status]),
			status,
		)
	}

	// Total shares counter
	ch <- prometheus.MustNewConstMetric(
		sharesCounterDesc,
		prometheus.GaugeValue,
		float64(totalShares),
	)
}
