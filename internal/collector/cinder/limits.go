package cinder

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/project"
	cinderdb "github.com/vexxhost/openstack_database_exporter/internal/db/cinder"
)

const (
	// Cinder default quota for gigabytes and backup_gigabytes
	cinderDefaultGigabytesQuota = 1000
)

var (
	limitsVolumeMaxGbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "limits_volume_max_gb"),
		"limits_volume_max_gb",
		[]string{
			"tenant",
			"tenant_id",
		},
		nil,
	)

	limitsVolumeUsedGbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "limits_volume_used_gb"),
		"limits_volume_used_gb",
		[]string{
			"tenant",
			"tenant_id",
		},
		nil,
	)

	limitsBackupMaxGbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "limits_backup_max_gb"),
		"limits_backup_max_gb",
		[]string{
			"tenant",
			"tenant_id",
		},
		nil,
	)

	limitsBackupUsedGbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "limits_backup_used_gb"),
		"limits_backup_used_gb",
		[]string{
			"tenant",
			"tenant_id",
		},
		nil,
	)

	volumeTypeQuotaGigabytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "volume_type_quota_gigabytes"),
		"volume_type_quota_gigabytes",
		[]string{
			"tenant",
			"tenant_id",
			"volume_type",
		},
		nil,
	)
)

type projectQuotaInfo struct {
	volumeMaxGB  int32
	volumeUsedGB int32
	backupMaxGB  int32
	backupUsedGB int32
	hasVolume    bool
	hasBackup    bool
}

type LimitsCollector struct {
	db              *sql.DB
	queries         *cinderdb.Queries
	logger          *slog.Logger
	projectResolver *project.Resolver
}

func NewLimitsCollector(db *sql.DB, logger *slog.Logger, projectResolver *project.Resolver) *LimitsCollector {
	return &LimitsCollector{
		db:      db,
		queries: cinderdb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "limits",
		),
		projectResolver: projectResolver,
	}
}

func (c *LimitsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- limitsVolumeMaxGbDesc
	ch <- limitsVolumeUsedGbDesc
	ch <- limitsBackupMaxGbDesc
	ch <- limitsBackupUsedGbDesc
	ch <- volumeTypeQuotaGigabytesDesc
}

func (c *LimitsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Get quota limits from cinder DB
	quotaLimits, err := c.queries.GetProjectQuotaLimits(ctx)
	if err != nil {
		c.logger.Error("failed to query quota limits", "error", err)
		return
	}

	// Get volume types for volume_type_quota_gigabytes
	volumeTypes, err := c.queries.GetVolumeTypes(ctx)
	if err != nil {
		c.logger.Error("failed to query volume types", "error", err)
		return
	}

	// Build per-project quota data from DB
	projectQuotas := make(map[string]*projectQuotaInfo)
	for _, quota := range quotaLimits {
		pid := quota.ProjectID.String
		if _, ok := projectQuotas[pid]; !ok {
			projectQuotas[pid] = &projectQuotaInfo{}
		}
		pq := projectQuotas[pid]

		switch quota.Resource {
		case "gigabytes":
			pq.volumeMaxGB = quota.HardLimit.Int32
			pq.volumeUsedGB = quota.InUse
			pq.hasVolume = true
		case "backup_gigabytes":
			pq.backupMaxGB = quota.HardLimit.Int32
			pq.backupUsedGB = quota.InUse
			pq.hasBackup = true
		}
	}

	// Build the full set of project IDs: union of DB projects and keystone projects
	allProjectIDs := make(map[string]string) // projectID -> projectName

	// Add projects from DB quotas (resolve name via keystone if available)
	for pid := range projectQuotas {
		name, _ := c.projectResolver.Resolve(pid)
		allProjectIDs[pid] = name
	}

	// Add projects from keystone that may not have explicit quotas
	for pid, info := range c.projectResolver.AllProjects() {
		if _, exists := allProjectIDs[pid]; !exists {
			allProjectIDs[pid] = info.Name
		}
	}

	// Emit metrics for all projects
	for projectID, projectName := range allProjectIDs {
		pq, hasExplicitQuota := projectQuotas[projectID]
		if !hasExplicitQuota {
			pq = &projectQuotaInfo{}
		}

		// Volume limits
		volumeMax := int32(cinderDefaultGigabytesQuota)
		if pq.hasVolume {
			volumeMax = pq.volumeMaxGB
		}
		ch <- prometheus.MustNewConstMetric(
			limitsVolumeMaxGbDesc,
			prometheus.GaugeValue,
			float64(volumeMax),
			projectName,
			projectID,
		)
		ch <- prometheus.MustNewConstMetric(
			limitsVolumeUsedGbDesc,
			prometheus.GaugeValue,
			float64(pq.volumeUsedGB),
			projectName,
			projectID,
		)

		// Backup limits (default 1000, 0 used)
		backupMax := int32(cinderDefaultGigabytesQuota)
		if pq.hasBackup {
			backupMax = pq.backupMaxGB
		}
		ch <- prometheus.MustNewConstMetric(
			limitsBackupMaxGbDesc,
			prometheus.GaugeValue,
			float64(backupMax),
			projectName,
			projectID,
		)
		ch <- prometheus.MustNewConstMetric(
			limitsBackupUsedGbDesc,
			prometheus.GaugeValue,
			float64(pq.backupUsedGB),
			projectName,
			projectID,
		)

		// Volume type quota gigabytes (default -1 per type)
		for _, vt := range volumeTypes {
			vtName := vt.Name.String

			// Check if there's an explicit per-type quota (e.g., "gigabytes_standard")
			perTypeResource := "gigabytes_" + vtName
			perTypeLimit := int32(-1) // default
			for _, quota := range quotaLimits {
				if quota.ProjectID.String == projectID && quota.Resource == perTypeResource {
					perTypeLimit = quota.HardLimit.Int32
					break
				}
			}

			ch <- prometheus.MustNewConstMetric(
				volumeTypeQuotaGigabytesDesc,
				prometheus.GaugeValue,
				float64(perTypeLimit),
				projectName,
				projectID,
				vtName,
			)
		}
	}
}
