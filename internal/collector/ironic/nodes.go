package ironic

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	ironicdb "github.com/vexxhost/openstack_database_exporter/internal/db/ironic"
)

const maxLabelLength = 128

func truncateLabel(s string) string {
	if len(s) > maxLabelLength {
		return s[:maxLabelLength]
	}
	return s
}

// NodesCollector collects per-node metrics for Ironic.
type NodesCollector struct {
	queries *ironicdb.Queries
	logger  *slog.Logger

	nodeMetric *prometheus.Desc
}

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

func (c *NodesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.nodeMetric
}

// CollectFromRows emits node metrics from pre-fetched rows.
func (c *NodesCollector) CollectFromRows(ch chan<- prometheus.Metric, nodes []ironicdb.GetNodeMetricsRow) {
	for _, node := range nodes {
		// Skip nodes with empty UUID to avoid duplicate label sets
		if !node.Uuid.Valid || node.Uuid.String == "" {
			c.logger.Debug("skipping node with empty UUID")
			continue
		}

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

		name := ""
		if node.Name.Valid {
			name = node.Name.String
		}

		metric, err := prometheus.NewConstMetric(
			c.nodeMetric,
			prometheus.GaugeValue,
			1,
			node.Uuid.String,
			name,
			powerState,
			provisionState,
			resourceClass,
			maintenance,
			consoleEnabled,
			retired,
			truncateLabel(node.RetiredReason),
		)
		if err != nil {
			c.logger.Error("failed to create node metric", "error", err)
			continue
		}
		ch <- metric
	}
}
