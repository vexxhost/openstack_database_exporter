package cinder

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
	cinderdb "github.com/vexxhost/openstack_database_exporter/internal/db/cinder"
)

var FAKE_TENANTS = map[string]string{
	"0c4e939acacf4376bdcd1129f1a054ad": "admin",
	"fdb8424c4e4f4c0ba32c52e2de3bd80e": "alt_demo",
	"0cbd49cbf76d405d9c86562e1d579bd3": "demo",
	"5961c443439d4fcebe42643723755e9d": "invisible_to_admin",
	"3d594eb0f04741069dbbb521635b21c7": "service",
	"43ebde53fc314b1c9ea2b8c5dc744927": "swifttenanttest1",
	"2db68fed84324f29bb73130c6c2094fb": "swifttenanttest2",
	"4b1eb781a47440acb8af9850103e537f": "swifttenanttest4",
}

var (
	limitsVolumeMaxGbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "limits_volume_max_gb"),
		"limits_volume_max_gb",
		[]string{
			"tenant",
			"tenant_id",
		},
		nil,
	)

	limitsVolumeUsedGbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "limits_volume_used_gb"),
		"limits_volume_used_gb",
		[]string{
			"tenant",
			"tenant_id",
		},
		nil,
	)

	limitsBackupMaxGbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "limits_backup_max_gb"),
		"limits_backup_max_gb",
		[]string{
			"tenant",
			"tenant_id",
		},
		nil,
	)

	limitsBackupUsedGbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "limits_backup_used_gb"),
		"limits_backup_used_gb",
		[]string{
			"tenant",
			"tenant_id",
		},
		nil,
	)
)

type LimitsCollector struct {
	db      *sql.DB
	queries *cinderdb.Queries
	logger  *slog.Logger
}

func NewLimitsCollector(db *sql.DB, logger *slog.Logger) *LimitsCollector {
	return &LimitsCollector{
		db:      db,
		queries: cinderdb.New(db),
		logger: logger.With(
			"namespace", collector.Namespace,
			"subsystem", Subsystem,
			"collector", "limits",
		),
	}
}

func (c *LimitsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- limitsVolumeMaxGbDesc
	ch <- limitsVolumeUsedGbDesc
	ch <- limitsBackupMaxGbDesc
	ch <- limitsBackupUsedGbDesc
}

func (c *LimitsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	quotaLimits, err := c.queries.GetProjectQuotaLimits(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return
	}

	for _, quota := range quotaLimits {
		// TODO(mnaser): Replace with tenant name when available
		projectName := quota.ProjectID.String
		if name, ok := FAKE_TENANTS[quota.ProjectID.String]; ok {
			projectName = name
		}

		switch quota.Resource {
		case "gigabytes":
			ch <- prometheus.MustNewConstMetric(
				limitsVolumeUsedGbDesc,
				prometheus.GaugeValue,
				float64(quota.InUse),
				projectName,
				quota.ProjectID.String,
			)
			ch <- prometheus.MustNewConstMetric(
				limitsVolumeMaxGbDesc,
				prometheus.GaugeValue,
				float64(quota.HardLimit.Int32),
				projectName,
				quota.ProjectID.String,
			)
		case "backup_gigabytes":
			ch <- prometheus.MustNewConstMetric(
				limitsBackupUsedGbDesc,
				prometheus.GaugeValue,
				float64(quota.InUse),
				projectName,
				quota.ProjectID.String,
			)
			ch <- prometheus.MustNewConstMetric(
				limitsBackupMaxGbDesc,
				prometheus.GaugeValue,
				float64(quota.HardLimit.Int32),
				projectName,
				quota.ProjectID.String,
			)
		default:
			c.logger.Warn("unknown quota resource", "resource", quota.Resource)
			continue
		}
	}
}
