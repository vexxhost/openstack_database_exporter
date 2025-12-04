package magnum

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
)

var (
	containerInfraUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)
)

type ContainerInfraCollector struct {
	db                *sql.DB
	logger            *slog.Logger
	clustersCollector *ClustersCollector
	mastersCollector  *MastersCollector
	nodesCollector    *NodesCollector
}

func NewContainerInfraCollector(db *sql.DB, logger *slog.Logger) *ContainerInfraCollector {
	return &ContainerInfraCollector{
		db:                db,
		logger:            logger,
		clustersCollector: NewClustersCollector(db, logger),
		mastersCollector:  NewMastersCollector(db, logger),
		nodesCollector:    NewNodesCollector(db, logger),
	}
}

func (c *ContainerInfraCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- containerInfraUpDesc
	c.clustersCollector.Describe(ch)
	c.mastersCollector.Describe(ch)
	c.nodesCollector.Describe(ch)
}

func (c *ContainerInfraCollector) Collect(ch chan<- prometheus.Metric) {
	var hasError bool

	// Collect metrics from all sub-collectors and track errors
	if err := c.clustersCollector.Collect(ch); err != nil {
		c.logger.Error("clusters collector failed", "error", err)
		hasError = true
	}

	if err := c.mastersCollector.Collect(ch); err != nil {
		c.logger.Error("masters collector failed", "error", err)
		hasError = true
	}

	if err := c.nodesCollector.Collect(ch); err != nil {
		c.logger.Error("nodes collector failed", "error", err)
		hasError = true
	}

	// Emit single up metric based on overall success
	upValue := float64(1)
	if hasError {
		upValue = float64(0)
	}
	ch <- prometheus.MustNewConstMetric(containerInfraUpDesc, prometheus.GaugeValue, upValue)
}
