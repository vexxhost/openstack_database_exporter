package nova

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	novadb "github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	novaapidb "github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
	placementdb "github.com/vexxhost/openstack_database_exporter/internal/db/placement"
)

var (
	novaUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)
)

type ComputeCollector struct {
	novaDB                *sql.DB
	novaApiDB             *sql.DB
	placementDB           *sql.DB
	logger                *slog.Logger
	servicesCollector     *ServicesCollector
	flavorsCollector      *FlavorsCollector
	instancesCollector    *InstancesCollector
	quotasCollector       *QuotasCollector
	limitsCollector       *LimitsCollector
	computeNodesCollector *ComputeNodesCollector
	serverCollector       *ServerCollector
}

func NewComputeCollector(novaDB, novaApiDB, placementDB *sql.DB, logger *slog.Logger) *ComputeCollector {
	novaQueries := novadb.New(novaDB)
	novaApiQueries := novaapidb.New(novaApiDB)

	var placementQueries *placementdb.Queries
	if placementDB != nil {
		placementQueries = placementdb.New(placementDB)
	}

	return &ComputeCollector{
		novaDB:                novaDB,
		novaApiDB:             novaApiDB,
		placementDB:           placementDB,
		logger:                logger,
		servicesCollector:     NewServicesCollector(logger, novaQueries, novaApiQueries),
		flavorsCollector:      NewFlavorsCollector(logger, novaQueries, novaApiQueries),
		instancesCollector:    NewInstancesCollector(logger, novaQueries, novaApiQueries),
		quotasCollector:       NewQuotasCollector(logger, novaQueries, novaApiQueries, placementQueries),
		limitsCollector:       NewLimitsCollector(logger, novaQueries, novaApiQueries, placementQueries),
		computeNodesCollector: NewComputeNodesCollector(logger, novaQueries, novaApiQueries),
		serverCollector:       NewServerCollector(logger, novaQueries, novaApiQueries),
	}
}

func (c *ComputeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- novaUpDesc
	c.servicesCollector.Describe(ch)
	c.flavorsCollector.Describe(ch)
	c.instancesCollector.Describe(ch)
	c.quotasCollector.Describe(ch)
	c.limitsCollector.Describe(ch)
	c.computeNodesCollector.Describe(ch)
	c.serverCollector.Describe(ch)
}

func (c *ComputeCollector) Collect(ch chan<- prometheus.Metric) {
	// Track if any sub-collector fails
	var hasError bool

	// Collect metrics from all sub-collectors
	if err := c.servicesCollector.Collect(ch); err != nil {
		c.logger.Error("Services collector failed", "error", err)
		hasError = true
	}

	if err := c.flavorsCollector.Collect(ch); err != nil {
		c.logger.Error("Flavors collector failed", "error", err)
		hasError = true
	}

	if err := c.instancesCollector.Collect(ch); err != nil {
		c.logger.Error("Instances collector failed", "error", err)
		hasError = true
	}

	if err := c.quotasCollector.Collect(ch); err != nil {
		c.logger.Error("Quotas collector failed", "error", err)
		hasError = true
	}

	if err := c.limitsCollector.Collect(ch); err != nil {
		c.logger.Error("Limits collector failed", "error", err)
		hasError = true
	}

	if err := c.computeNodesCollector.Collect(ch); err != nil {
		c.logger.Error("Compute nodes collector failed", "error", err)
		hasError = true
	}

	if err := c.serverCollector.Collect(ch); err != nil {
		c.logger.Error("Server collector failed", "error", err)
		hasError = true
	}

	// Emit single up metric based on overall success/failure
	upValue := float64(1)
	if hasError {
		upValue = 0
	}
	ch <- prometheus.MustNewConstMetric(
		novaUpDesc,
		prometheus.GaugeValue,
		upValue,
	)
}
