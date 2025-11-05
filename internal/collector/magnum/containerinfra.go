package magnum

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
	magnumdb "github.com/vexxhost/openstack_database_exporter/internal/db/magnum"
)

const Subsystem = "container_infra"

var (
	// Known cluster statuses from the original openstack-exporter
	clusterStatuses = []string{
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

	magnumUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)

	clusterMastersDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "cluster_masters"),
		"cluster_masters",
		[]string{
			"uuid",
			"name",
			"stack_id",
			"status",
			"node_count",
			"project_id",
		},
		nil,
	)

	clusterNodesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "cluster_nodes"),
		"cluster_nodes",
		[]string{
			"uuid",
			"name",
			"stack_id",
			"status",
			"master_count",
			"project_id",
		},
		nil,
	)

	clusterStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "cluster_status"),
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

	totalClustersDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "total_clusters"),
		"total_clusters",
		nil,
		nil,
	)
)

type ContainerInfraCollector struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewContainerInfraCollector(db *sql.DB, logger *slog.Logger) *ContainerInfraCollector {
	return &ContainerInfraCollector{
		db:     db,
		logger: logger,
	}
}

func (c *ContainerInfraCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- magnumUpDesc
	ch <- clusterMastersDesc
	ch <- clusterNodesDesc
	ch <- clusterStatusDesc
	ch <- totalClustersDesc
}

func (c *ContainerInfraCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	queries := magnumdb.New(c.db)

	// Set up metric = 1 (service is up)
	ch <- prometheus.MustNewConstMetric(
		magnumUpDesc,
		prometheus.GaugeValue,
		1,
	)

	// Collect cluster metrics
	c.collectClusterMetrics(ctx, queries, ch)
}

func (c *ContainerInfraCollector) collectClusterMetrics(ctx context.Context, queries *magnumdb.Queries, ch chan<- prometheus.Metric) {
	clusters, err := queries.GetClusterMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to get cluster metrics", "error", err)
		return
	}

	// total_clusters count
	ch <- prometheus.MustNewConstMetric(
		totalClustersDesc,
		prometheus.GaugeValue,
		float64(len(clusters)),
	)

	// Individual cluster metrics
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

		masterCountStr := fmt.Sprintf("%d", masterCount)
		nodeCountStr := fmt.Sprintf("%d", nodeCount)

		// cluster_masters metric
		ch <- prometheus.MustNewConstMetric(
			clusterMastersDesc,
			prometheus.GaugeValue,
			float64(masterCount),
			uuid,
			name,
			cluster.StackID,
			cluster.Status,
			nodeCountStr,
			projectID,
		)

		// cluster_nodes metric
		ch <- prometheus.MustNewConstMetric(
			clusterNodesDesc,
			prometheus.GaugeValue,
			float64(nodeCount),
			uuid,
			name,
			cluster.StackID,
			cluster.Status,
			masterCountStr,
			projectID,
		)

		// cluster_status metric
		statusValue := mapClusterStatus(cluster.Status)
		ch <- prometheus.MustNewConstMetric(
			clusterStatusDesc,
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

func mapClusterStatus(status string) int {
	for idx, s := range clusterStatuses {
		if status == s {
			return idx
		}
	}
	return -1
}
