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
	// Agent state metrics
	agentStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "agent_state"),
		"agent_state",
		[]string{"adminState", "agent_version", "availability_zone", "binary", "host", "project"},
		nil,
	)

	// Service count metrics
	servicesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "services"),
		"services",
		nil,
		nil,
	)

	// Service information metrics
	serviceInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "service_info"),
		"Nova service information",
		[]string{"id", "uuid", "host", "binary", "topic", "disabled", "forced_down"},
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
	ch <- servicesDesc
	ch <- serviceInfoDesc
}

func (c *ServicesCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	services, err := c.novaDB.GetServices(ctx)
	if err != nil {
		return fmt.Errorf("failed to get services: %w", err)
	}

	// Emit service count
	ch <- prometheus.MustNewConstMetric(
		servicesDesc,
		prometheus.GaugeValue,
		float64(len(services)),
	)

	// Emit per-service metrics
	for _, service := range services {
		// Agent state metric (1 = up, 0 = down based on last_seen_up and disabled status)
		adminState := "enabled"
		if service.Disabled.Valid && service.Disabled.Bool {
			adminState = "disabled"
		}

		agentValue := float64(1) // Assume up unless we have specific down indicators
		if (service.Disabled.Valid && service.Disabled.Bool) || (service.ForcedDown.Valid && service.ForcedDown.Bool) {
			agentValue = 0
		}

		ch <- prometheus.MustNewConstMetric(
			agentStateDesc,
			prometheus.GaugeValue,
			agentValue,
			adminState,
			fmt.Sprintf("%d", service.Version.Int32),
			"nova", // Default availability zone for Nova
			nullStringToString(service.Binary),
			nullStringToString(service.Host),
			"nova", // Project name
		)

		// Service information metric
		ch <- prometheus.MustNewConstMetric(
			serviceInfoDesc,
			prometheus.GaugeValue,
			1, // Info metric always has value 1
			fmt.Sprintf("%d", service.ID),
			nullStringToString(service.Uuid),
			nullStringToString(service.Host),
			nullStringToString(service.Binary),
			nullStringToString(service.Topic),
			fmt.Sprintf("%t", service.Disabled.Valid && service.Disabled.Bool),
			fmt.Sprintf("%t", service.ForcedDown.Valid && service.ForcedDown.Bool),
		)
	}

	return nil
}
