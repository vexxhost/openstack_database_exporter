package octavia

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
	octaviadb "github.com/vexxhost/openstack_database_exporter/internal/db/octavia"
)

var (
	amphoraStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "amphora_status"),
		"amphora_status",
		[]string{
			"id",
			"loadbalancer_id",
			"compute_id",
			"status",
			"role",
			"lb_network_ip",
			"ha_ip",
			"cert_expiration",
		},
		nil,
	)

	totalAmphoraeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "total_amphorae"),
		"total_amphorae",
		nil,
		nil,
	)
)

type AmphoraCollector struct {
	db      *sql.DB
	queries *octaviadb.Queries
	logger  *slog.Logger
}

func NewAmphoraCollector(db *sql.DB, logger *slog.Logger) *AmphoraCollector {
	return &AmphoraCollector{
		db:      db,
		queries: octaviadb.New(db),
		logger: logger.With(
			"namespace", collector.Namespace,
			"subsystem", Subsystem,
			"collector", "amphora",
		),
	}
}

func (c *AmphoraCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- amphoraStatusDesc
	ch <- totalAmphoraeDesc
}

func (c *AmphoraCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	amphorae, err := c.queries.GetAllAmphora(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return
	}

	for _, amphora := range amphorae {
		certExp := ""
		if amphora.CertExpiration.Valid {
			certExp = amphora.CertExpiration.Time.Format(time.RFC3339)
		}

		ch <- prometheus.MustNewConstMetric(
			amphoraStatusDesc,
			prometheus.GaugeValue,
			statusToValue(amphora.Status, []string{
				"BOOTING",
				"ALLOCATED",
				"READY",
				"PENDING_CREATE",
				"PENDING_DELETE",
				"DELETED",
				"ERROR",
			}),
			amphora.ID,
			amphora.LoadBalancerID.String,
			amphora.ComputeID.String,
			amphora.Status,
			amphora.Role.String,
			amphora.LbNetworkIp.String,
			amphora.HaIp.String,
			certExp,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		totalAmphoraeDesc,
		prometheus.GaugeValue,
		float64(len(amphorae)),
	)
}
