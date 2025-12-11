package magnum

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	magnumdb "github.com/vexxhost/openstack_database_exporter/internal/db/magnum"
)

var (
	// Known cluster statuses from the original openstack-exporter
	knownClusterStatuses = []string{
		"CREATE_COMPLETE",
		"CREATE_FAILED",
		"CREATE_IN_PROGRESS",
		"UPDATE_IN_PROGRESS",
		"UPDATE_FAILED",
		"UPDATE_COMPLETE",
		"DELETE_IN_PROGRESS",
		"DELETE_FAILED",
		"DELETE_COMPLETE",
		"RESUME_COMPLETE",
		"RESUME_FAILED",
		"RESTORE_COMPLETE",
		"ROLLBACK_IN_PROGRESS",
		"ROLLBACK_FAILED",
		"ROLLBACK_COMPLETE",
		"SNAPSHOT_COMPLETE",
		"CHECK_COMPLETE",
		"ADOPT_COMPLETE",
	}

	clustersStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "cluster_status"),
		"cluster_status",
		[]string{
			"uuid",
			"name",
			"stack_id",
			"status",
			"node_count",
			"master_count",
			"project_id",
		},
		nil,
	)

	clustersCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "total_clusters"),
		"total_clusters",
		nil,
		nil,
	)
)

type ClustersCollector struct {
	db      *sql.DB
	queries *magnumdb.Queries
	logger  *slog.Logger
}

func NewClustersCollector(db *sql.DB, logger *slog.Logger) *ClustersCollector {
	return &ClustersCollector{
		db:      db,
		queries: magnumdb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "clusters",
		),
	}
}

func (c *ClustersCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- clustersStatusDesc
	ch <- clustersCountDesc
}

func (c *ClustersCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	clusters, err := c.queries.GetClusterMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to get cluster metrics", "error", err)
		return
	}

	// total_clusters count
	ch <- prometheus.MustNewConstMetric(
		clustersCountDesc,
		prometheus.GaugeValue,
		float64(len(clusters)),
	)

	// Individual cluster status metrics
	for _, cluster := range clusters {
		uuid := ""
		if cluster.Uuid.Valid {
			uuid = cluster.Uuid.String
		}

		name := ""
		if cluster.Name.Valid {
			name = cluster.Name.String
		}

		projectID := ""
		if cluster.ProjectID.Valid {
			projectID = cluster.ProjectID.String
		}

		// Convert interface{} to int for counts
		masterCount := 0
		if cluster.MasterCount != nil {
			if mc, ok := cluster.MasterCount.(int64); ok {
				masterCount = int(mc)
			}
		}

		nodeCount := 0
		if cluster.NodeCount != nil {
			if nc, ok := cluster.NodeCount.(int64); ok {
				nodeCount = int(nc)
			}
		}

		masterCountStr := formatCount(masterCount)
		nodeCountStr := formatCount(nodeCount)

		// cluster_status metric
		statusValue := mapClusterStatusValue(cluster.Status)
		ch <- prometheus.MustNewConstMetric(
			clustersStatusDesc,
			prometheus.GaugeValue,
			float64(statusValue),
			uuid,
			name,
			cluster.StackID,
			cluster.Status,
			nodeCountStr,
			masterCountStr,
			projectID,
		)
	}
}

func mapClusterStatusValue(status string) int {
	for idx, s := range knownClusterStatuses {
		if status == s {
			return idx
		}
	}
	return -1
}

func formatCount(count int) string {
	return fmt.Sprintf("%d", count)
}
