package manila

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	maniladb "github.com/vexxhost/openstack_database_exporter/internal/db/manila"
	"github.com/vexxhost/openstack_database_exporter/internal/util"
)

// volumeStatuses matches the upstream openstack-exporter mapVolumeStatus list exactly.
// The upstream Manila exporter reuses the Cinder volume status mapping for share_status.
var volumeStatuses = []string{
	"creating",
	"available",
	"attaching",
	"detaching",
	"in-use",
	"maintenance",
	"deleting",
	"awaiting-transfer",
	"error",
	"error_deleting",
	"backing-up",
	"restoring-backup",
	"error_backing-up",
	"error_restoring",
	"error_extending",
	"downloading",
	"uploading",
	"retyping",
	"extending",
}

var (
	// Known share statuses from the original openstack-exporter
	shareStatuses = []string{
		"available", "creating", "deleting", "error", "error_deleting",
		"extending", "inactive", "managing", "migrating", "migration_error",
		"restoring", "reverting", "reverting_error", "reverting_to_snapshot",
		"shrinking", "shrinking_error", "soft_deleting", "unmanaging", "updating",
	}

	manilaUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)

	shareGbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "share_gb"),
		"share_gb",
		[]string{
			"id",
			"name",
			"status",
			"availability_zone",
			"share_type",
			"share_proto",
			"share_type_name",
			"project_id",
		},
		nil,
	)

	shareStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "share_status"),
		"share_status",
		[]string{
			"id",
			"name",
			"status",
			"size",
			"share_type",
			"share_proto",
			"share_type_name",
			"project_id",
		},
		nil,
	)

	shareStatusCounterDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "share_status_counter"),
		"share_status_counter",
		[]string{
			"status",
		},
		nil,
	)

	sharesCounterDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "shares_counter"),
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
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "shares",
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
		shareType := share.ShareType
		shareTypeName := share.ShareTypeName
		availabilityZone := share.AvailabilityZone

		totalShares++

		// Count status for counter metrics
		if status != "" {
			statusCounts[status]++
		}

		// share_gb metric - size in GB per share
		// Label order matches upstream: id, name, status, availability_zone, share_type, share_proto, share_type_name, project_id
		ch <- prometheus.MustNewConstMetric(
			shareGbDesc,
			prometheus.GaugeValue,
			float64(size),
			shareID,
			name,
			status,
			availabilityZone,
			shareType,
			shareProto,
			shareTypeName,
			projectID,
		)

		// share_status metric - uses mapVolumeStatus like upstream openstack-exporter
		// Label order matches upstream: id, name, status, size, share_type, share_proto, share_type_name, project_id
		sizeStr := "0"
		if share.Size.Valid {
			sizeStr = fmt.Sprintf("%d", share.Size.Int32)
		}

		ch <- prometheus.MustNewConstMetric(
			shareStatusDesc,
			prometheus.GaugeValue,
			util.StatusToValue(status, volumeStatuses),
			shareID,
			name,
			status,
			sizeStr,
			shareType,
			shareProto,
			shareTypeName,
			projectID,
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
