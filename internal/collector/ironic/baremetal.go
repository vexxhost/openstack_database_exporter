package ironic

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	ironicdb "github.com/vexxhost/openstack_database_exporter/internal/db/ironic"
)

// BaremetalCollector is the umbrella collector for Ironic baremetal metrics.
// It queries the database once and passes the result to NodesCollector.
type BaremetalCollector struct {
	queries *ironicdb.Queries
	logger  *slog.Logger

	upMetric       *prometheus.Desc
	nodesCollector *NodesCollector
}

func NewBaremetalCollector(db *sql.DB, logger *slog.Logger) *BaremetalCollector {
	return &BaremetalCollector{
		queries: ironicdb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "baremetal",
		),

		upMetric: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, Subsystem, "up"),
			"Whether the Ironic baremetal service is up",
			nil,
			nil,
		),

		nodesCollector: NewNodesCollector(db, logger),
	}
}

func (c *BaremetalCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upMetric
	c.nodesCollector.Describe(ch)
}

func (c *BaremetalCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Query node metrics once and reuse for the nodes sub-collector
	nodes, err := c.queries.GetNodeMetrics(ctx)
	if err != nil {
		c.logger.Error("failed to query Ironic database", "error", err)
		ch <- prometheus.MustNewConstMetric(c.upMetric, prometheus.GaugeValue, 0)
		return
	}

	c.nodesCollector.CollectFromRows(ch, nodes)

	ch <- prometheus.MustNewConstMetric(c.upMetric, prometheus.GaugeValue, 1)
}
