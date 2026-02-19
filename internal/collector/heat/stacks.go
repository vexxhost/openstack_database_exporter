package heat

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	heatdb "github.com/vexxhost/openstack_database_exporter/internal/db/heat"
)

var (
	// Known stack statuses from the original openstack-exporter
	knownStackStatuses = []string{
		"INIT_IN_PROGRESS",
		"INIT_FAILED",
		"INIT_COMPLETE",
		"CREATE_IN_PROGRESS",
		"CREATE_FAILED",
		"CREATE_COMPLETE",
		"DELETE_IN_PROGRESS",
		"DELETE_FAILED",
		"DELETE_COMPLETE",
		"UPDATE_IN_PROGRESS",
		"UPDATE_FAILED",
		"UPDATE_COMPLETE",
		"ROLLBACK_IN_PROGRESS",
		"ROLLBACK_FAILED",
		"ROLLBACK_COMPLETE",
		"SUSPEND_IN_PROGRESS",
		"SUSPEND_FAILED",
		"SUSPEND_COMPLETE",
		"RESUME_IN_PROGRESS",
		"RESUME_FAILED",
		"RESUME_COMPLETE",
		"ADOPT_IN_PROGRESS",
		"ADOPT_FAILED",
		"ADOPT_COMPLETE",
		"SNAPSHOT_IN_PROGRESS",
		"SNAPSHOT_FAILED",
		"SNAPSHOT_COMPLETE",
		"CHECK_IN_PROGRESS",
		"CHECK_FAILED",
		"CHECK_COMPLETE",
	}

	stacksUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)

	stackStatusCounterDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "stack_status_counter"),
		"stack_status_counter",
		[]string{
			"status",
		},
		nil,
	)
)

type StacksCollector struct {
	queries *heatdb.Queries
	logger  *slog.Logger
}

func NewStacksCollector(db *sql.DB, logger *slog.Logger) *StacksCollector {
	return &StacksCollector{
		queries: heatdb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "stacks",
		),
	}
}

func (c *StacksCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- stacksUpDesc
	ch <- stackStatusCounterDesc
}

func (c *StacksCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	stacks, err := c.queries.GetStackMetrics(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(stacksUpDesc, prometheus.GaugeValue, 0)
		c.logger.Error("failed to query stacks", "error", err)
		return
	}

	// Initialize status counters
	stackStatusCounter := make(map[string]int, len(knownStackStatuses))
	for _, status := range knownStackStatuses {
		stackStatusCounter[status] = 0
	}

	// Count status occurrences
	for _, stack := range stacks {
		if _, ok := stackStatusCounter[stack.Status]; ok {
			stackStatusCounter[stack.Status]++
		}
	}

	// Stack status counter metrics in stable order
	for _, status := range knownStackStatuses {
		ch <- prometheus.MustNewConstMetric(
			stackStatusCounterDesc,
			prometheus.GaugeValue,
			float64(stackStatusCounter[status]),
			status,
		)
	}

	ch <- prometheus.MustNewConstMetric(stacksUpDesc, prometheus.GaugeValue, 1)
}
