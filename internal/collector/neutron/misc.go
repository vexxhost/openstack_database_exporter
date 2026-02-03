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
	networkIPAvailabilitiesUsed = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "network_ip_availabilities_used"),
		"network_ip_availabilities_used",
		[]string{
			"network_id",
			"subnet_id",
			"project_id",
			"network_name",
			"subnet_name",
			"cidr",
			"ip_version",
		},
		nil,
	)
)

type MiscCollector struct {
	db      *sql.DB
	queries *neutrondb.Queries
	logger  *slog.Logger
}

func NewMiscCollector(db *sql.DB, logger *slog.Logger) *MiscCollector {
	return &MiscCollector{
		db:      db,
		queries: neutrondb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "subnets",
		),
	}
}

func (c *MiscCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- networkIPAvailabilitiesUsed
}

func (c *MiscCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	nipaus, err := c.queries.GetNetworkIPAvailabilitiesUsed(ctx)
	if err != nil {
		c.logger.Error("failed to query", "error", err)
		return err
	}

	for _, nipau := range nipaus {
		ch <- prometheus.MustNewConstMetric(
			networkIPAvailabilitiesUsed,
			prometheus.GaugeValue,
			cast.ToFloat64(nipau.AllocationCount),
			nipau.NetworkID,
			nipau.SubnetID,
			nipau.ProjectID.String,
			nipau.SubnetName.String,
			nipau.NetworkName.String,
			nipau.Cidr.String,
			cast.ToString(nipau.IpVersion.Int32),
		)
	}
	return nil
}
