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

var (
	nodesUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)

	clusterNodesCountDesc = prometheus.NewDesc(
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
)

type NodesCollector struct {
	db      *sql.DB
	queries *magnumdb.Queries
	logger  *slog.Logger
}

func NewNodesCollector(db *sql.DB, logger *slog.Logger) *NodesCollector {
	return &NodesCollector{
		db:      db,
		queries: magnumdb.New(db),
		logger: logger.With(
			"namespace", collector.Namespace,
			"subsystem", Subsystem,
			"collector", "nodes",
		),
	}
}

func (c *NodesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- nodesUpDesc
	ch <- clusterNodesCountDesc
}

func (c *NodesCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	clusters, err := c.queries.GetClusterMetrics(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(nodesUpDesc, prometheus.GaugeValue, 0)
		c.logger.Error("Failed to get cluster metrics for nodes", "error", err)
		return
	}

	// Individual cluster node metrics
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

	ch <- prometheus.MustNewConstMetric(nodesUpDesc, prometheus.GaugeValue, 1)
}
