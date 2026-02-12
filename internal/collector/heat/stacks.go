package heat

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

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

	// stackStatusValueMap provides O(1) lookup for status -> index mapping.
	// Values are explicitly assigned to keep metric encodings stable even
	// if knownStackStatuses is reordered or modified.
	stackStatusValueMap = map[string]int{
		"INIT_IN_PROGRESS":     0,
		"INIT_FAILED":          1,
		"INIT_COMPLETE":        2,
		"CREATE_IN_PROGRESS":   3,
		"CREATE_FAILED":        4,
		"CREATE_COMPLETE":      5,
		"DELETE_IN_PROGRESS":   6,
		"DELETE_FAILED":        7,
		"DELETE_COMPLETE":      8,
		"UPDATE_IN_PROGRESS":   9,
		"UPDATE_FAILED":        10,
		"UPDATE_COMPLETE":      11,
		"ROLLBACK_IN_PROGRESS": 12,
		"ROLLBACK_FAILED":      13,
		"ROLLBACK_COMPLETE":    14,
		"SUSPEND_IN_PROGRESS":  15,
		"SUSPEND_FAILED":       16,
		"SUSPEND_COMPLETE":     17,
		"RESUME_IN_PROGRESS":   18,
		"RESUME_FAILED":        19,
		"RESUME_COMPLETE":      20,
		"ADOPT_IN_PROGRESS":    21,
		"ADOPT_FAILED":         22,
		"ADOPT_COMPLETE":       23,
		"SNAPSHOT_IN_PROGRESS": 24,
		"SNAPSHOT_FAILED":      25,
		"SNAPSHOT_COMPLETE":    26,
		"CHECK_IN_PROGRESS":    27,
		"CHECK_FAILED":         28,
		"CHECK_COMPLETE":       29,
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
	ch <- stackStatusDesc
	ch <- stackStatusCounterDesc
	ch <- totalStacksDesc
}

func (c *StacksCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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
		// Count status occurrences only for known statuses
		if _, ok := stackStatusCounter[stack.Status]; ok {
			stackStatusCounter[stack.Status]++
		}

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

	// Stack status counter metrics in stable order
	for _, status := range knownStackStatuses {
		ch <- prometheus.MustNewConstMetric(
			stackStatusCounterDesc,
			prometheus.GaugeValue,
			float64(stackStatusCounter[status]),
			status,
		)
	}

	return nil
}

func mapStackStatusValue(status string) int {
	if v, ok := stackStatusValueMap[status]; ok {
		return v
	}
	return -1
}
