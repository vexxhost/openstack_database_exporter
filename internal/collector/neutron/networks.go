package neutron

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

var (
	networkDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "network"),
		"network",
		[]string{
			"id",
			"name",
			"project_id",
			"provider_network_type",
			"provider_physical_network",
			"provider_segmentation_id",
			"status",
			"subnets",
			"is_external",
			"is_shared",
		},
		nil,
	)
	networksDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "networks"),
		"networks",
		nil,
		nil,
	)
)

type NetworkCollector struct {
	db      *sql.DB
	queries *neutrondb.Queries
	logger  *slog.Logger
}

func NewNetworkCollector(db *sql.DB, logger *slog.Logger) *NetworkCollector {
	return &NetworkCollector{
		db:      db,
		queries: neutrondb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "networks",
		),
	}
}

func (c *NetworkCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- networkDesc
}

func (c *NetworkCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	nets, err := c.queries.GetNetworks(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		networksDesc,
		prometheus.GaugeValue,
		float64(len(nets)),
	)

	for _, net := range nets {
		ch <- prometheus.MustNewConstMetric(
			networkDesc,
			prometheus.GaugeValue,
			cast.ToFloat64(1),
			net.ID,
			net.Name.String,
			net.ProjectID.String,
			net.ProviderNetworkType.String,
			net.ProviderPhysicalNetwork.String,
			cast.ToString(net.ProviderSegmentationID.Int32),
			net.Status.String,
			cast.ToString(net.Subnets),
			cast.ToString(cast.ToBool(net.IsExternal)),
			cast.ToString(cast.ToBool(net.IsShared)),
		)
	}
	return nil
}
