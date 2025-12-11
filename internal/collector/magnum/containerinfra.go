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
	// Collect metrics from all sub-collectors
	c.clustersCollector.Collect(ch)
	c.mastersCollector.Collect(ch)
	c.nodesCollector.Collect(ch)

	// Emit up metric (individual collectors handle their own error logging)
	ch <- prometheus.MustNewConstMetric(containerInfraUpDesc, prometheus.GaugeValue, 1)
}
