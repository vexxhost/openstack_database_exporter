package heat

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	orchestrationUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)
)

type OrchestrationCollector struct {
	logger          *slog.Logger
	stacksCollector *StacksCollector
}

func NewOrchestrationCollector(db *sql.DB, logger *slog.Logger) *OrchestrationCollector {
	return &OrchestrationCollector{
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "orchestration",
		),
		stacksCollector: NewStacksCollector(db, logger),
	}
}

func (c *OrchestrationCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- orchestrationUpDesc
	c.stacksCollector.Describe(ch)
}

func (c *OrchestrationCollector) Collect(ch chan<- prometheus.Metric) {
	var hasError bool

	// Collect metrics from stacks collector and track errors
	if err := c.stacksCollector.Collect(ch); err != nil {
		hasError = true
	}

	// Emit single up metric based on overall success
	upValue := float64(1)
	if hasError {
		upValue = float64(0)
	}
	ch <- prometheus.MustNewConstMetric(orchestrationUpDesc, prometheus.GaugeValue, upValue)
}
