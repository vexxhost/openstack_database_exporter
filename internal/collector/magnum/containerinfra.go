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
	containerInfraUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)
)

// ContainerInfraCollector is a single collector that queries the database once
// and emits all magnum/container_infra metrics: total_clusters, cluster_status,
// cluster_masters, and cluster_nodes.
type ContainerInfraCollector struct {
	db      *sql.DB
	queries *magnumdb.Queries
	logger  *slog.Logger
}

func NewContainerInfraCollector(db *sql.DB, logger *slog.Logger) *ContainerInfraCollector {
	return &ContainerInfraCollector{
		db:      db,
		queries: magnumdb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "container_infra",
		),
	}
}

func (c *ContainerInfraCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- containerInfraUpDesc
	ch <- clustersStatusDesc
	ch <- clustersCountDesc
	ch <- clusterMastersCountDesc
	ch <- clusterNodesCountDesc
}

func (c *ContainerInfraCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	clusters, err := c.queries.GetClusterMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to get cluster metrics", "error", err)
		ch <- prometheus.MustNewConstMetric(containerInfraUpDesc, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(containerInfraUpDesc, prometheus.GaugeValue, 1)

	// total_clusters count
	ch <- prometheus.MustNewConstMetric(
		clustersCountDesc,
		prometheus.GaugeValue,
		float64(len(clusters)),
	)

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

		masterCount := int(cluster.MasterCount)
		nodeCount := int(cluster.NodeCount)

		masterCountStr := fmt.Sprintf("%d", masterCount)
		nodeCountStr := fmt.Sprintf("%d", nodeCount)

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

		// cluster_masters metric
		ch <- prometheus.MustNewConstMetric(
			clusterMastersCountDesc,
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
			clusterNodesCountDesc,
			prometheus.GaugeValue,
			float64(nodeCount),
			uuid,
			name,
			cluster.StackID,
			cluster.Status,
			masterCountStr,
			projectID,
		)
	}
}
