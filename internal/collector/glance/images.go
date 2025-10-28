package glance

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
	glancedb "github.com/vexxhost/openstack_database_exporter/internal/db/glance"
)

var (
	imagesUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)

	imagesBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "image_bytes"),
		"image_bytes",
		[]string{
			"id",
			"name",
			"tenant_id",
		},
		nil,
	)

	imagesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "images"),
		"images",
		nil,
		nil,
	)
)

type ImagesCollector struct {
	db      *sql.DB
	queries *glancedb.Queries
	logger  *slog.Logger
}

func NewImagesCollector(db *sql.DB, logger *slog.Logger) *ImagesCollector {
	return &ImagesCollector{
		db:      db,
		queries: glancedb.New(db),
		logger: logger.With(
			"namespace", collector.Namespace,
			"subsystem", Subsystem,
			"collector", "images",
		),
	}
}

func (c *ImagesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- imagesUpDesc
	ch <- imagesBytesDesc
	ch <- imagesDesc
}

func (c *ImagesCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	images, err := c.queries.GetAllImages(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(imagesUpDesc, prometheus.GaugeValue, 0)

		c.logger.Error("failed to query", "error", err)
		return
	}

	for _, image := range images {
		// Convert size from nullable int64 to float64, defaulting to 0 if null
		sizeBytes := float64(0)
		if image.Size.Valid {
			sizeBytes = float64(image.Size.Int64)
		}

		ch <- prometheus.MustNewConstMetric(
			imagesBytesDesc,
			prometheus.GaugeValue,
			sizeBytes,
			image.ID,
			image.Name.String,
			image.Owner.String,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		imagesDesc,
		prometheus.GaugeValue,
		float64(len(images)),
	)

	ch <- prometheus.MustNewConstMetric(imagesUpDesc, prometheus.GaugeValue, 1)
}