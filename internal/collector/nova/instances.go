package nova

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	"github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
)

// InstancesCollector collects metrics about Nova instances
type InstancesCollector struct {
	logger          *slog.Logger
	novaDB          *nova.Queries
	novaAPIDB       *nova_api.Queries
	instanceMetrics map[string]*prometheus.Desc
}

// NewInstancesCollector creates a new instances collector
func NewInstancesCollector(logger *slog.Logger, novaDB *nova.Queries, novaAPIDB *nova_api.Queries) *InstancesCollector {
	return &InstancesCollector{
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "instances",
		),
		novaDB:    novaDB,
		novaAPIDB: novaAPIDB,
		instanceMetrics: map[string]*prometheus.Desc{
			"instance": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "instance"),
				"Nova instance information",
				[]string{"id", "uuid", "name", "user_id", "project_id", "host", "availability_zone", "vm_state", "power_state", "task_state"},
				nil,
			),
		},
	}
}

// Describe implements the prometheus.Collector interface
func (c *InstancesCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.instanceMetrics {
		ch <- desc
	}
}

// Collect implements the prometheus.Collector interface
func (c *InstancesCollector) Collect(ch chan<- prometheus.Metric) error {
	return c.collectInstanceMetrics(ch)
}

func (c *InstancesCollector) collectInstanceMetrics(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	instances, err := c.novaDB.GetInstances(ctx)
	if err != nil {
		return err
	}

	for _, instance := range instances {
		// Convert fields to strings for labels
		var (
			id               = fmt.Sprintf("%d", instance.ID)
			uuid             = instance.Uuid // Already a string
			name             = nullStringToString(instance.DisplayName)
			userID           = nullStringToString(instance.UserID)
			projectID        = nullStringToString(instance.ProjectID)
			host             = nullStringToString(instance.Host)
			availabilityZone = nullStringToString(instance.AvailabilityZone)
			vmState          = nullStringToString(instance.VmState)
			powerState       = nullInt32ToString(instance.PowerState)
			taskState        = nullStringToString(instance.TaskState)
		)

		ch <- prometheus.MustNewConstMetric(
			c.instanceMetrics["instance"],
			prometheus.GaugeValue,
			1,
			id, uuid, name, userID, projectID, host, availabilityZone, vmState, powerState, taskState,
		)
	}

	return nil
}

// Helper functions for converting nullable SQL types
func nullInt32ToString(ni sql.NullInt32) string {
	if ni.Valid {
		return fmt.Sprintf("%d", ni.Int32)
	}
	return ""
}
