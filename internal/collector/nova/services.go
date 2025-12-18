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
}

func (c *ServicesCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	services, err := c.novaDB.GetServices(ctx)
	if err != nil {
		return fmt.Errorf("failed to get services: %w", err)
	}

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

		// Determine zone based on service binary (matching original logic)
		zone := "nova" // Default zone for compute services
		if service.Binary.Valid && service.Binary.String == "nova-scheduler" {
			zone = "internal"
		}

		ch <- prometheus.MustNewConstMetric(
			agentStateDesc,
			prometheus.CounterValue, // Original uses counter, not gauge
			agentValue,
			adminState,
			disabledReason,
			nullStringToString(service.Host),
			fmt.Sprintf("%d", service.ID),
			nullStringToString(service.Binary),
			zone,
		)
	}

	return nil
}
