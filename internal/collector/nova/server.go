package nova

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	"github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
)

var (
	// Known server statuses from the original openstack-exporter
	knownServerStatuses = []string{
		"ACTIVE",            // The server is active.
		"BUILDING",          // The server has not yet finished the initial boot process.
		"DELETED",           // The server is deleted.
		"ERROR",             // The server is in error.
		"HARD_REBOOT",       // The server is hard rebooting.
		"PASSWORD",          // The password is being reset on the server.
		"REBOOT",            // The server is in a soft reboot state.
		"REBUILD",           // The server is currently being rebuilt from an image.
		"RESCUE",            // The server is in rescue mode.
		"RESIZE",            // Server is performing the differential copy of data that changed during its initial copy.
		"SHUTOFF",           // The virtual machine (VM) was powered down by the user, but not through the OpenStack Compute API.
		"SUSPENDED",         // The server is suspended, either by request or necessity.
		"UNKNOWN",           // The state of the server is unknown. Contact your cloud provider.
		"VERIFY_RESIZE",     // System is awaiting confirmation that the server is operational after a move or resize.
		"MIGRATING",         // The server is migrating. This is caused by a live migration (moving a server that is active) action.
		"PAUSED",            // The server is paused.
		"REVERT_RESIZE",     // The resize or migration of a server failed for some reason. The destination server is being cleaned up and the original source server is restarting.
		"SHELVED",           // The server is in shelved state. Depends on the shelve offload time, the server will be automatically shelved off loaded.
		"SHELVED_OFFLOADED", // The shelved server is offloaded (removed from the compute host) and it needs unshelved action to be used again.
		"SOFT_DELETED",      // The server is marked as deleted but will remain in the cloud for some configurable amount of time.
	}
)

// ServerCollector collects metrics about Nova servers (instances)
type ServerCollector struct {
	logger        *slog.Logger
	novaDB        *nova.Queries
	novaAPIDB     *nova_api.Queries
	serverMetrics map[string]*prometheus.Desc
}

// NewServerCollector creates a new server collector
func NewServerCollector(logger *slog.Logger, novaDB *nova.Queries, novaAPIDB *nova_api.Queries) *ServerCollector {
	return &ServerCollector{
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "server",
		),
		novaDB:    novaDB,
		novaAPIDB: novaAPIDB,
		serverMetrics: map[string]*prometheus.Desc{
			"server_local_gb": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "server_local_gb"),
				"server_local_gb",
				[]string{"id", "name", "tenant_id"},
				nil,
			),
			"server_status": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "server_status"),
				"server_status",
				[]string{"address_ipv4", "address_ipv6", "availability_zone", "flavor_id", "host_id", "hypervisor_hostname", "id", "instance_libvirt", "name", "status", "tenant_id", "user_id", "uuid"},
				nil,
			),
			"total_vms": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "total_vms"),
				"total_vms",
				nil,
				nil,
			),
			"availability_zones": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "availability_zones"),
				"availability_zones",
				nil,
				nil,
			),
		},
	}
}

// Describe implements the prometheus.Collector interface
func (c *ServerCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.serverMetrics {
		ch <- desc
	}
}

// Collect implements the prometheus.Collector interface
func (c *ServerCollector) Collect(ch chan<- prometheus.Metric) error {
	return c.collectServerMetrics(ch)
}

func (c *ServerCollector) collectServerMetrics(ch chan<- prometheus.Metric) error {
	instances, err := c.novaDB.GetInstances(context.Background())
	if err != nil {
		return err
	}

	// Count total VMs and availability zones
	totalVMs := len(instances)
	azSet := make(map[string]bool)

	for _, instance := range instances {
		if instance.AvailabilityZone.Valid && instance.AvailabilityZone.String != "" {
			azSet[instance.AvailabilityZone.String] = true
		}

		// Server local GB - using root_gb from instance
		ch <- prometheus.MustNewConstMetric(
			c.serverMetrics["server_local_gb"],
			prometheus.GaugeValue,
			float64(instance.RootGb.Int32),
			instance.Uuid,
			instance.DisplayName.String,
			instance.ProjectID.String,
		)

		// Server status - detailed instance information using proper status mapping
		statusValue := float64(mapServerStatus(instance.VmState.String))

		// Build instance name for libvirt
		var instanceLibvirt string
		instanceLibvirt = fmt.Sprintf("instance-%08x", instance.ID)

		ch <- prometheus.MustNewConstMetric(
			c.serverMetrics["server_status"],
			prometheus.GaugeValue,
			statusValue,
			"", // address_ipv4 - would need separate query for fixed IPs
			"", // address_ipv6 - would need separate query for fixed IPs
			instance.AvailabilityZone.String,
			strconv.FormatInt(int64(instance.InstanceTypeID.Int32), 10),
			instance.Host.String,
			instance.Host.String, // hypervisor_hostname same as host in simple setups
			strconv.FormatInt(int64(instance.ID), 10),
			instanceLibvirt,
			instance.DisplayName.String,
			instance.VmState.String,
			instance.ProjectID.String,
			instance.UserID.String,
			instance.Uuid,
		)
	}

	// Total VMs count
	ch <- prometheus.MustNewConstMetric(
		c.serverMetrics["total_vms"],
		prometheus.GaugeValue,
		float64(totalVMs),
	)

	// Availability zones count
	ch <- prometheus.MustNewConstMetric(
		c.serverMetrics["availability_zones"],
		prometheus.GaugeValue,
		float64(len(azSet)),
	)

	return nil
}

func mapServerStatus(status string) int {
	for idx, s := range knownServerStatuses {
		if status == s {
			return idx
		}
	}
	return -1
}
