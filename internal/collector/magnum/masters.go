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
	clusterMastersCountDesc = prometheus.NewDesc(
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
)

type MastersCollector struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewMastersCollector(db *sql.DB, logger *slog.Logger) *MastersCollector {
	return &MastersCollector{
		db:     db,
		logger: logger,
	}
}

func (c *MastersCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- clusterMastersCountDesc
}

func (c *MastersCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	queries := magnumdb.New(c.db)

	c.collectMasterMetrics(ctx, queries, ch)
}

func (c *MastersCollector) collectMasterMetrics(ctx context.Context, queries *magnumdb.Queries, ch chan<- prometheus.Metric) {
	clusters, err := queries.GetClusterMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to get cluster metrics for masters", "error", err)
		return
	}

	// Individual cluster master metrics
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

		nodeCountStr := fmt.Sprintf("%d", nodeCount)

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
	}
}
