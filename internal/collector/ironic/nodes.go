package ironic

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	ironicdb "github.com/vexxhost/openstack_database_exporter/internal/db/ironic"
)

// maxLabelLength is the maximum length for free-form label values to prevent
// unbounded cardinality and memory usage in Prometheus.
const maxLabelLength = 128

// truncateLabel truncates a string to maxLabelLength.
func truncateLabel(s string) string {
	if len(s) > maxLabelLength {
		return s[:maxLabelLength]
	}
	return s
}

// NodesCollector collects metrics about Ironic nodes
type NodesCollector struct {
	queries *ironicdb.Queries
	logger  *slog.Logger

	// Individual node metrics
	nodeMetric *prometheus.Desc
}

// NewNodesCollector creates a new NodesCollector
func NewNodesCollector(db *sql.DB, logger *slog.Logger) *NodesCollector {
	return &NodesCollector{
		queries: ironicdb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "nodes",
		),

		nodeMetric: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, Subsystem, "node"),
			"Ironic node status",
			[]string{
				"id", "name", "power_state", "provision_state",
				"resource_class", "maintenance", "console_enabled", "retired", "retired_reason",
			},
			nil,
		),
	}
}

// Describe implements prometheus.Collector
func (c *NodesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.nodeMetric
}

// Collect queries the database and collects node metrics into the provided channel.
func (c *NodesCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	nodes, err := c.queries.GetNodeMetrics(ctx)
	if err != nil {
		c.logger.Error("failed to get node metrics", "error", err)
		return err
	}

	return c.CollectFromRows(ch, nodes)
}

// CollectFromRows emits node metrics from pre-fetched rows.
func (c *NodesCollector) CollectFromRows(ch chan<- prometheus.Metric, nodes []ironicdb.GetNodeMetricsRow) error {
	for _, node := range nodes {
		// Skip nodes with empty UUID to avoid duplicate label sets
		if !node.Uuid.Valid || node.Uuid.String == "" {
			nodeName := ""
			if node.Name.Valid {
				nodeName = node.Name.String
			}
			c.logger.Debug("skipping node with empty UUID", "name", nodeName)
			continue
		}
		// Individual node status metric
		maintenance := "false"
		if node.Maintenance.Valid && node.Maintenance.Bool {
			maintenance = "true"
		}

		powerState := "unknown"
		if node.PowerState.Valid {
			powerState = node.PowerState.String
		}

		provisionState := "unknown"
		if node.ProvisionState.Valid {
			provisionState = node.ProvisionState.String
		}

		resourceClass := "unknown"
		if node.ResourceClass.Valid {
			resourceClass = node.ResourceClass.String
		}

		consoleEnabled := "false"
		if node.ConsoleEnabled.Valid && node.ConsoleEnabled.Bool {
			consoleEnabled = "true"
		}

		retired := "false"
		if node.Retired.Valid && node.Retired.Bool {
			retired = "true"
		}

		retiredReason := truncateLabel(node.RetiredReason)

		name := ""
		if node.Name.Valid {
			name = node.Name.String
		}

		nodeUUID := node.Uuid.String

		// Emit individual node metric
		metric, err := prometheus.NewConstMetric(
			c.nodeMetric,
			prometheus.GaugeValue,
			1,
			nodeUUID, name, powerState, provisionState, resourceClass, maintenance, consoleEnabled, retired, retiredReason,
		)
		if err != nil {
			c.logger.Error("failed to create node metric", "error", err)
			continue
		}
		ch <- metric
	}

	return nil
}
