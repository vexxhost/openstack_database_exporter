package util

import "github.com/prometheus/client_golang/prometheus"

// downCollector is a Prometheus collector that emits a single _up=0 gauge.
// Used when a service is configured but the database connection fails at startup.
type downCollector struct {
	upDesc *prometheus.Desc
}

// NewDownCollector creates a collector that always reports _up=0 for a service
// whose database URL was configured but the connection could not be established.
func NewDownCollector(namespace, subsystem string) prometheus.Collector {
	return &downCollector{
		upDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "up"),
			"up",
			nil,
			nil,
		),
	}
}

func (c *downCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upDesc
}

func (c *downCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.upDesc, prometheus.GaugeValue, 0)
}
