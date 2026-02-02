package neutron

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

var (
	haRouterAgentPortBindingDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "l3_agent_of_router"),
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
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "ha_router_agent_port_bindings",
		),
	}
}

func (c *HARouterAgentPortBindingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- haRouterAgentPortBindingDesc
}

func (c *HARouterAgentPortBindingCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	bindings, err := c.queries.GetHARouterAgentPortBindingsWithAgents(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return err
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
	return nil
}

var (
	routerDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "router"),
		"router",
		[]string{
			"id",
			"name",
			"status",
			"admin_state_up",
			"project_id",
			"external_network_id",
		},
		nil,
	)
	routersDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "routers"),
		"routers",
		nil,
		nil,
	)
	notActiveRoutersDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "routers_not_active"),
		"routers_not_active",
		nil,
		nil,
	)
)

type RouterCollector struct {
	db      *sql.DB
	queries *neutrondb.Queries
	logger  *slog.Logger
}

func NewRouterCollector(db *sql.DB, logger *slog.Logger) *RouterCollector {
	return &RouterCollector{
		db:      db,
		queries: neutrondb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "routers",
		),
	}
}

func (c *RouterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- routerDesc
	ch <- routersDesc
	ch <- notActiveRoutersDesc
}

func (c *RouterCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	routers, err := c.queries.GetRouters(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return err
	}

	naRouters := 0

	ch <- prometheus.MustNewConstMetric(
		routersDesc,
		prometheus.GaugeValue,
		float64(len(routers)),
	)

	for _, router := range routers {
		if router.Status.String != "ACTIVE" {
			naRouters += 1
		}
		ch <- prometheus.MustNewConstMetric(
			routerDesc,
			prometheus.GaugeValue,
			cast.ToFloat64(1),
			router.ID,
			router.Name.String,
			router.Status.String,
			cast.ToString(router.AdminStateUp.Bool),
			router.ProjectID.String,
			router.GwPortID.String,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		notActiveRoutersDesc,
		prometheus.GaugeValue,
		float64(naRouters),
	)
	return nil

}
