package masakari

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	masakaridb "github.com/vexxhost/openstack_database_exporter/internal/db/masakari"
)

var (
	masakariUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)

	segmentDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "segment"),
		"segment",
		[]string{"id", "uuid", "name", "description", "recovery_method", "service_type"},
		nil,
	)

	hostDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "host"),
		"host",
		[]string{"id", "uuid", "hostname", "failover_segment_id", "failover_segment_name", "type", "control_attributes"},
		nil,
	)

	hostOnMaintenanceDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "host_on_maintenance"),
		"host_on_maintenance",
		[]string{"uuid", "hostname", "failover_segment_id"},
		nil,
	)

	hostReservedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "host_reserved"),
		"host_reserved",
		[]string{"uuid", "hostname", "failover_segment_id"},
		nil,
	)

	notificationsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "notifications"),
		"notifications",
		[]string{"status"},
		nil,
	)
)

type SegmentsCollector struct {
	db      *sql.DB
	queries *masakaridb.Queries
	logger  *slog.Logger
}

func NewSegmentsCollector(db *sql.DB, logger *slog.Logger) *SegmentsCollector {
	return &SegmentsCollector{
		db:      db,
		queries: masakaridb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "segments",
		),
	}
}

func (c *SegmentsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- masakariUpDesc
	ch <- segmentDesc
	ch <- hostDesc
	ch <- hostOnMaintenanceDesc
	ch <- hostReservedDesc
	ch <- notificationsDesc
}

func (c *SegmentsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	segments, err := c.queries.GetFailoverSegments(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(masakariUpDesc, prometheus.GaugeValue, 0)
		c.logger.Error("failed to query failover segments", "error", err)
		return
	}

	for _, seg := range segments {
		description := ""
		if seg.Description.Valid {
			description = seg.Description.String
		}

		ch <- prometheus.MustNewConstMetric(
			segmentDesc,
			prometheus.GaugeValue, 1.0,
			strconv.Itoa(int(seg.ID)),
			seg.Uuid,
			seg.Name,
			description,
			string(seg.RecoveryMethod),
			seg.ServiceType,
		)
	}

	hosts, err := c.queries.GetHosts(ctx)
	if err != nil {
		c.logger.Error("failed to query hosts", "error", err)
	} else {
		for _, host := range hosts {
			segmentName := ""
			if host.FailoverSegmentName.Valid {
				segmentName = host.FailoverSegmentName.String
			}

			ch <- prometheus.MustNewConstMetric(
				hostDesc,
				prometheus.GaugeValue, 1.0,
				strconv.Itoa(int(host.ID)),
				host.Uuid,
				host.Name,
				host.FailoverSegmentID,
				segmentName,
				host.Type,
				host.ControlAttributes,
			)

			var onMaintenance float64
			if host.OnMaintenance {
				onMaintenance = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				hostOnMaintenanceDesc,
				prometheus.GaugeValue, onMaintenance,
				host.Uuid, host.Name, host.FailoverSegmentID,
			)

			var reserved float64
			if host.Reserved {
				reserved = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				hostReservedDesc,
				prometheus.GaugeValue, reserved,
				host.Uuid, host.Name, host.FailoverSegmentID,
			)
		}
	}

	notifications, err := c.queries.GetNotificationsByStatus(ctx)
	if err != nil {
		c.logger.Error("failed to query notifications", "error", err)
	} else {
		for _, n := range notifications {
			ch <- prometheus.MustNewConstMetric(
				notificationsDesc,
				prometheus.GaugeValue, float64(n.Count),
				string(n.Status),
			)
		}
	}

	ch <- prometheus.MustNewConstMetric(masakariUpDesc, prometheus.GaugeValue, 1)
}
