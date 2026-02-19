package neutron

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

var (
	networkDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "network"),
		"network",
		[]string{
			"id",
			"is_external",
			"is_shared",
			"name",
			"provider_network_type",
			"provider_physical_network",
			"provider_segmentation_id",
			"status",
			"subnets",
			"tags",
			"tenant_id",
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
	ch <- networksDesc
}

func (c *NetworkCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	networks, err := c.queries.GetNetworks(ctx)
	if err != nil {
		c.logger.Error("failed to query networks", "error", err)
		return
	}

	for _, n := range networks {
		ch <- prometheus.MustNewConstMetric(
			networkDesc,
			prometheus.GaugeValue,
			0,
			n.ID,
			strconv.FormatBool(n.IsExternal == 1),
			strconv.FormatBool(n.IsShared == 1),
			n.Name.String,
			n.ProviderNetworkType.String,
			n.ProviderPhysicalNetwork.String,
			dbString(n.ProviderSegmentationID),
			n.Status.String,
			dbString(n.Subnets),
			dbString(n.Tags),
			n.ProjectID.String,
		)
	}

	ch <- prometheus.MustNewConstMetric(networksDesc, prometheus.GaugeValue, float64(len(networks)))
}

// dbString converts a sqlc interface{} value (from COALESCE/GROUP_CONCAT) to string.
// MySQL driver returns []byte for CHAR/VARCHAR columns scanned into interface{}.
func dbString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case []byte:
		return string(s)
	case string:
		return s
	default:
		return fmt.Sprintf("%v", v)
	}
}
