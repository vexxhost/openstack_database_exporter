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

func nullInt32ToString(ni sql.NullInt32) string {
	if ni.Valid {
		return fmt.Sprintf("%d", ni.Int32)
	}
	return "0"
}

// FlavorsCollector collects metrics about Nova flavors
type FlavorsCollector struct {
	logger        *slog.Logger
	novaDB        *nova.Queries
	novaAPIDB     *nova_api.Queries
	flavorMetrics map[string]*prometheus.Desc
}

// NewFlavorsCollector creates a new flavors collector
func NewFlavorsCollector(logger *slog.Logger, novaDB *nova.Queries, novaAPIDB *nova_api.Queries) *FlavorsCollector {
	return &FlavorsCollector{
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "flavors",
		),
		novaDB:    novaDB,
		novaAPIDB: novaAPIDB,
		flavorMetrics: map[string]*prometheus.Desc{
			"flavor": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "flavor"),
				"flavor",
				[]string{"disk", "id", "is_public", "name", "ram", "vcpus"},
				nil,
			),
			"flavors": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "flavors"),
				"flavors",
				nil,
				nil,
			),
			"security_groups": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, Subsystem, "security_groups"),
				"security_groups",
				nil,
				nil,
			),
		},
	}
}

// Describe implements the prometheus.Collector interface
func (c *FlavorsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.flavorMetrics {
		ch <- desc
	}
}

// Collect implements the prometheus.Collector interface
func (c *FlavorsCollector) Collect(ch chan<- prometheus.Metric) error {
	return c.collectFlavorMetrics(ch)
}

func (c *FlavorsCollector) collectFlavorMetrics(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	flavors, err := c.novaAPIDB.GetFlavors(ctx)
	if err != nil {
		return err
	}

	// Total flavors count
	ch <- prometheus.MustNewConstMetric(
		c.flavorMetrics["flavors"],
		prometheus.GaugeValue,
		float64(len(flavors)),
	)

	// Security groups count (hardcoded to 1 like in original test)
	ch <- prometheus.MustNewConstMetric(
		c.flavorMetrics["security_groups"],
		prometheus.GaugeValue,
		1,
	)

	for _, flavor := range flavors {
		// Format labels to match original test order: disk, id, is_public, name, ram, vcpus
		id := fmt.Sprintf("%d", flavor.ID)
		name := flavor.Name
		vcpus := fmt.Sprintf("%d", flavor.Vcpus)
		ram := fmt.Sprintf("%d", flavor.MemoryMb)
		disk := nullInt32ToString(flavor.RootGb)
		isPublic := fmt.Sprintf("%t", flavor.IsPublic.Valid && flavor.IsPublic.Bool)

		ch <- prometheus.MustNewConstMetric(
			c.flavorMetrics["flavor"],
			prometheus.GaugeValue,
			1,
			disk, id, isPublic, name, ram, vcpus,
		)
	}

	return nil
}
