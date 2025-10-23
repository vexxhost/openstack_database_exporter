package neutron

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

var (
	haRouterAgentPortBindingDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "l3_agent_of_router"),
		"l3_agent_of_router",
		[]string{
			"router_id",
			"l3_agent_id",
			"ha_state",
			"agent_alive",
			"agent_admin_up",
			"agent_host",
		},
		nil,
	)
)

type HARouterAgentPortBindingCollector struct {
	db      *sql.DB
	queries *neutrondb.Queries
	logger  *slog.Logger
}

func NewHARouterAgentPortBindingCollector(db *sql.DB, logger *slog.Logger) *HARouterAgentPortBindingCollector {
	return &HARouterAgentPortBindingCollector{
		db:      db,
		queries: neutrondb.New(db),
		logger: logger.With(
			"namespace", collector.Namespace,
			"subsystem", Subsystem,
			"collector", "ha_router_agent_port_bindings",
		),
	}
}

func (c *HARouterAgentPortBindingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- haRouterAgentPortBindingDesc
}

func (c *HARouterAgentPortBindingCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	bindings, err := c.queries.GetHARouterAgentPortBindingsWithAgents(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return
	}

	for _, binding := range bindings {
		if !binding.L3AgentID.Valid {
			continue
		}

		alive := time.Since(binding.AgentHeartbeatTimestamp.Time).Seconds() < 75

		ch <- prometheus.MustNewConstMetric(
			haRouterAgentPortBindingDesc,
			prometheus.GaugeValue,
			cast.ToFloat64(alive),
			binding.RouterID,
			binding.L3AgentID.String,
			string(binding.State.HaRouterAgentPortBindingsState),
			strconv.FormatBool(alive),
			strconv.FormatBool(binding.AgentAdminStateUp.Bool),
			binding.AgentHost.String,
		)
	}
}
