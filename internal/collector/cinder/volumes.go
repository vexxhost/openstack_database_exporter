package cinder

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
	cinderdb "github.com/vexxhost/openstack_database_exporter/internal/db/cinder"
)

var (
	volumesUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)

	volumeGbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "volume_gb"),
		"volume_gb",
		[]string{
			"id",
			"name",
			"status",
			"availability_zone",
			"bootable",
			"tenant_id",
			"user_id",
			"volume_type",
			"server_id",
		},
		nil,
	)

	volumeStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "volume_status"),
		"volume_status",
		[]string{
			"id",
			"name",
			"status",
			"bootable",
			"tenant_id",
			"size",
			"volume_type",
			"server_id",
		},
		nil,
	)

	volumeStatusCounterDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "volume_status_counter"),
		"volume_status_counter",
		[]string{
			"status",
		},
		nil,
	)

	volumesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "volumes"),
		"volumes",
		nil,
		nil,
	)
)

type VolumesCollector struct {
	db      *sql.DB
	queries *cinderdb.Queries
	logger  *slog.Logger
}

func NewVolumesCollector(db *sql.DB, logger *slog.Logger) *VolumesCollector {
	return &VolumesCollector{
		db:      db,
		queries: cinderdb.New(db),
		logger: logger.With(
			"namespace", collector.Namespace,
			"subsystem", Subsystem,
			"collector", "volumes",
		),
	}
}

func (c *VolumesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- volumesDesc
	ch <- volumeGbDesc
	ch <- volumeStatusDesc
	ch <- volumeStatusCounterDesc
	ch <- volumesUpDesc
}

func (c *VolumesCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	volumes, err := c.queries.GetAllVolumes(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(volumesUpDesc, prometheus.GaugeValue, 0)

		c.logger.Error("failed to query", "error", err)
		return
	}

	volume_status_counter := map[string]int{
		"creating":          0,
		"available":         0,
		"reserved":          0,
		"attaching":         0,
		"detaching":         0,
		"in-use":            0,
		"maintenance":       0,
		"deleting":          0,
		"awaiting-transfer": 0,
		"error":             0,
		"error_deleting":    0,
		"backing-up":        0,
		"restoring-backup":  0,
		"error_backing-up":  0,
		"error_restoring":   0,
		"error_extending":   0,
		"downloading":       0,
		"uploading":         0,
		"retyping":          0,
		"extending":         0,
	}

	for _, volume := range volumes {
		volume_status_counter[volume.Status.String]++

		ch <- prometheus.MustNewConstMetric(
			volumeGbDesc,
			prometheus.GaugeValue,
			float64(volume.Size.Int32),
			volume.ID,
			volume.Name.String,
			volume.Status.String,
			volume.AvailabilityZone.String,
			strconv.FormatBool(volume.Bootable.Bool),
			volume.ProjectID.String,
			volume.UserID.String,
			volume.VolumeType.String,
			volume.ServerID.String,
		)

		ch <- prometheus.MustNewConstMetric(
			volumeStatusDesc,
			prometheus.GaugeValue,
			collector.StatusToValue(volume.Status.String, []string{
				"creating",
				"available",
				"reserved",
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
			}),
			volume.ID,
			volume.Name.String,
			volume.Status.String,
			strconv.FormatBool(volume.Bootable.Bool),
			volume.ProjectID.String,
			strconv.Itoa(int(volume.Size.Int32)),
			volume.VolumeType.String,
			volume.ServerID.String,
		)
	}

	for status, count := range volume_status_counter {
		ch <- prometheus.MustNewConstMetric(
			volumeStatusCounterDesc,
			prometheus.GaugeValue,
			float64(count),
			status,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		volumesDesc,
		prometheus.GaugeValue,
		float64(len(volumes)),
	)

	ch <- prometheus.MustNewConstMetric(volumesUpDesc, prometheus.GaugeValue, 1)
}
