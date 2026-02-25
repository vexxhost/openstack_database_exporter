package neutron

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

var (
	routerDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "router"),
		"router",
		[]string{
			"admin_state_up",
			"external_network_id",
			"id",
			"name",
			"project_id",
			"status",
		},
		nil,
	)

	routersDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "routers"),
		"routers",
		nil,
		nil,
	)

	routersNotActiveDesc = prometheus.NewDesc(
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
	ch <- routersNotActiveDesc
}

func (c *RouterCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	routers, err := c.queries.GetRouters(ctx)
	if err != nil {
		c.logger.Error("failed to query routers", "error", err)
		return
	}

	notActive := 0
	for _, r := range routers {
		ch <- prometheus.MustNewConstMetric(
			routerDesc,
			prometheus.GaugeValue,
			1,
			strconv.FormatBool(r.AdminStateUp.Bool),
			r.ExternalNetworkID,
			r.ID,
			r.Name.String,
			r.ProjectID.String,
			r.Status.String,
		)

		if r.Status.String != "ACTIVE" {
			notActive++
		}
	}

	ch <- prometheus.MustNewConstMetric(routersDesc, prometheus.GaugeValue, float64(len(routers)))
	ch <- prometheus.MustNewConstMetric(routersNotActiveDesc, prometheus.GaugeValue, float64(notActive))
}
