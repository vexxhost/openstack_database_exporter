package ironic

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	ironicdb "github.com/vexxhost/openstack_database_exporter/internal/db/ironic"
)

// BaremetalCollector is the umbrella collector for Ironic baremetal metrics
type BaremetalCollector struct {
	db      *sql.DB
	queries *ironicdb.Queries
	logger  *slog.Logger

	// Single up metric for baremetal service
	upMetric *prometheus.Desc

	// Sub-collectors
	nodesCollector *NodesCollector
}

// NewBaremetalCollector creates a new umbrella collector for Ironic baremetal service
func NewBaremetalCollector(db *sql.DB, logger *slog.Logger) *BaremetalCollector {
	return &BaremetalCollector{
		db:      db,
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

		// Initialize sub-collectors
		nodesCollector: NewNodesCollector(db, logger),
	}
}

// Describe implements prometheus.Collector
func (c *BaremetalCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upMetric
	c.nodesCollector.Describe(ch)
}

// Collect implements prometheus.Collector
func (c *BaremetalCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	up := float64(1)

	// Test database connectivity by running a simple query
	_, err := c.queries.GetNodeMetrics(ctx)
	if err != nil {
		c.logger.Error("failed to query Ironic database", "error", err)
		up = 0
	}

	// Emit up metric
	upMetric, err := prometheus.NewConstMetric(
		c.upMetric,
		prometheus.GaugeValue,
		up,
	)
	if err != nil {
		c.logger.Error("failed to create up metric", "error", err)
	} else {
		ch <- upMetric
	}

	// Only collect from sub-collectors if we're up
	if up == 1 {
		// Collect nodes metrics
		if err := c.nodesCollector.Collect(ch); err != nil {
			c.logger.Error("nodes collector failed", "error", err)
		}
	}
}
