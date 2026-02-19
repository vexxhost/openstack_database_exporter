package neutron

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

var (
	agentStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "agent_state"),
		"agent_state",
		[]string{
			"id",
			"hostname",
			"service",
			"adminState",
			"zone",
		},
		nil,
	)
)

type AgentsCollector struct {
	db      *sql.DB
	queries *neutrondb.Queries
	logger  *slog.Logger
}

func NewAgentsCollector(db *sql.DB, logger *slog.Logger) *AgentsCollector {
	return &AgentsCollector{
		db:      db,
		queries: neutrondb.New(db),
		logger: logger.With(
			"namespace", Namespace,
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

	agents, err := c.queries.GetAgents(ctx)
	if err != nil {
		c.logger.Error("failed to query agents", "error", err)
		return
	}

	for _, agent := range agents {
		ch <- prometheus.MustNewConstMetric(
			agentStateDesc,
			prometheus.GaugeValue,
			float64(agent.Alive),
			agent.ID,
			agent.Hostname,
			agent.Service,
			agent.AdminState,
			agent.Zone.String,
		)
	}
}
