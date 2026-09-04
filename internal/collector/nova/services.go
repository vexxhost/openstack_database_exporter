package nova

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	novadb "github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	novaapidb "github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
)

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

var (
	// Agent state metrics - matches original openstack-exporter
	agentStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "agent_state"),
		"agent_state",
		[]string{"adminState", "disabledReason", "hostname", "id", "service", "zone"},
		nil,
	)

	availabilityZonesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "availability_zones"),
		"availability_zones",
		nil,
		nil,
	)
)

type ServicesCollector struct {
	logger    *slog.Logger
	novaDB    *novadb.Queries
	novaAPIDB *novaapidb.Queries
}

func NewServicesCollector(logger *slog.Logger, novaDB *novadb.Queries, novaAPIDB *novaapidb.Queries) *ServicesCollector {
	return &ServicesCollector{
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "services",
		),
		novaDB:    novaDB,
		novaAPIDB: novaAPIDB,
	}
}

func (c *ServicesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- agentStateDesc
	ch <- availabilityZonesDesc
}

func (c *ServicesCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	services, err := c.novaDB.GetServices(ctx)
	if err != nil {
		return fmt.Errorf("failed to get services: %w", err)
	}

	// Build host → availability zone map from aggregate metadata
	hostAZMap := make(map[string]string)
	azRows, err := c.novaAPIDB.GetHostAvailabilityZones(ctx)
	if err != nil {
		c.logger.Error("failed to get host availability zones", "error", err)
	} else {
		for _, row := range azRows {
			if row.Host.Valid && row.AvailabilityZone.Valid {
				hostAZMap[row.Host.String] = row.AvailabilityZone.String
			}
		}
	}

	// Count distinct availability zones from aggregate metadata
	azSet := make(map[string]struct{})
	for _, az := range hostAZMap {
		if az != "" {
			azSet[az] = struct{}{}
		}
	}
	ch <- prometheus.MustNewConstMetric(availabilityZonesDesc, prometheus.GaugeValue, float64(len(azSet)))

	// Emit per-service agent state metrics matching original exporter
	for _, service := range services {
		// Determine admin state and disabled reason
		adminState := "enabled"
		disabledReason := ""
		agentValue := float64(1) // 1 for enabled, 0 for disabled

		if service.Disabled.Valid && service.Disabled.Bool {
			adminState = "disabled"
			agentValue = 0
			if service.DisabledReason.Valid {
				disabledReason = service.DisabledReason.String
			}
		}

		// Determine zone: use AZ from aggregate metadata if available,
		// otherwise default to "internal" for non-compute services
		zone := "internal"
		if service.Binary.Valid && service.Binary.String == "nova-compute" {
			zone = "nova"
		}
		if service.Host.Valid {
			if az, ok := hostAZMap[service.Host.String]; ok {
				zone = az
			}
		}

		ch <- prometheus.MustNewConstMetric(
			agentStateDesc,
			prometheus.GaugeValue,
			agentValue,
			adminState,
			disabledReason,
			nullStringToString(service.Host),
			nullStringToString(service.Uuid),
			nullStringToString(service.Binary),
			zone,
		)
	}

	return nil
}
