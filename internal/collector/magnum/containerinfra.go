package magnum

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
)

var (
	magnumUpDesc = prometheus.NewDesc(
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
	ch <- magnumUpDesc
	c.clustersCollector.Describe(ch)
	c.mastersCollector.Describe(ch)
	c.nodesCollector.Describe(ch)
}

func (c *ContainerInfraCollector) Collect(ch chan<- prometheus.Metric) {
	// Set up metric = 1 (service is up)
	ch <- prometheus.MustNewConstMetric(
		magnumUpDesc,
		prometheus.GaugeValue,
		1,
	)

	// Collect metrics from all sub-collectors
	c.clustersCollector.Collect(ch)
	c.mastersCollector.Collect(ch)
	c.nodesCollector.Collect(ch)
}
