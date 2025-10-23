package cinder

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
	cinderdb "github.com/vexxhost/openstack_database_exporter/internal/db/cinder"
)

var (
	agentStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "agent_state"),
		"agent_state",
		[]string{
			"uuid",
			"hostname",
			"service",
			"adminState",
			"zone",
			"disabledReason",
		},
		nil,
	)
)

type AgentsCollector struct {
	db      *sql.DB
	queries *cinderdb.Queries
	logger  *slog.Logger
}

func NewAgentsCollector(db *sql.DB, logger *slog.Logger) *AgentsCollector {
	return &AgentsCollector{
		db:      db,
		queries: cinderdb.New(db),
		logger: logger.With(
			"namespace", collector.Namespace,
			"subsystem", Subsystem,
			"collector", "agents",
		),
	}
}

func (c *AgentsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- agentStateDesc
}

func (c *AgentsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	services, err := c.queries.GetAllServices(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return
	}

	for _, service := range services {
		ch <- prometheus.MustNewConstMetric(
			agentStateDesc,
			prometheus.GaugeValue,
			float64(service.State),
			service.Uuid.String,
			service.Host.String,
			service.Service.String,
			service.AdminState,
			service.Zone.String,
			service.DisabledReason.String,
		)
	}
}
