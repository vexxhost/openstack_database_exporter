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

	stackStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "stack_status"),
		"stack_status",
		[]string{
			"id",
			"name",
			"project_id",
			"status",
			"action",
		},
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

	totalStacksDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "total_stacks"),
		"total_stacks",
		nil,
		nil,
	)
)

type StacksCollector struct {
	db      *sql.DB
	queries *heatdb.Queries
	logger  *slog.Logger
}

func NewStacksCollector(db *sql.DB, logger *slog.Logger) *StacksCollector {
	return &StacksCollector{
		db:      db,
		queries: heatdb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "stacks",
		),
	}
}

func (c *StacksCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- stackStatusDesc
	ch <- stackStatusCounterDesc
	ch <- totalStacksDesc
}

func (c *StacksCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	stacks, err := c.queries.GetStackMetrics(ctx)
	if err != nil {
		c.logger.Error("failed to query stacks", "error", err)
		return err
	}

	// Initialize status counters
	stackStatusCounter := make(map[string]int, len(knownStackStatuses))
	for _, status := range knownStackStatuses {
		stackStatusCounter[status] = 0
	}

	// total_stacks count
	ch <- prometheus.MustNewConstMetric(
		totalStacksDesc,
		prometheus.GaugeValue,
		float64(len(stacks)),
	)

	// Individual stack status metrics and count status occurrences
	for _, stack := range stacks {
		// Count status occurrences
		stackStatusCounter[stack.Status]++

		// stack_status metric
		statusValue := mapStackStatusValue(stack.Status)
		ch <- prometheus.MustNewConstMetric(
			stackStatusDesc,
			prometheus.GaugeValue,
			float64(statusValue),
			stack.ID,
			stack.Name,
			stack.Tenant,
			stack.Status,
			stack.Action,
		)
	}

	// Stack status counter metrics
	for status, count := range stackStatusCounter {
		ch <- prometheus.MustNewConstMetric(
			stackStatusCounterDesc,
			prometheus.GaugeValue,
			float64(count),
			status,
		)
	}

	return nil
}

func mapStackStatusValue(status string) int {
	for idx, s := range knownStackStatuses {
		if status == s {
			return idx
		}
	}
	return -1
}
