package keystone

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector"
)

var (
	keystoneUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)
)

type IdentityCollector struct {
	db                *sql.DB
	logger            *slog.Logger
	domainsCollector  *DomainsCollector
	projectsCollector *ProjectsCollector
	groupsCollector   *GroupsCollector
	regionsCollector  *RegionsCollector
	usersCollector    *UsersCollector
}

func NewIdentityCollector(db *sql.DB, logger *slog.Logger) *IdentityCollector {
	return &IdentityCollector{
		db:                db,
		logger:            logger,
		domainsCollector:  NewDomainsCollector(db, logger),
		projectsCollector: NewProjectsCollector(db, logger),
		groupsCollector:   NewGroupsCollector(db, logger),
		regionsCollector:  NewRegionsCollector(db, logger),
		usersCollector:    NewUsersCollector(db, logger),
	}
}

func (c *IdentityCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- keystoneUpDesc
	c.domainsCollector.Describe(ch)
	c.projectsCollector.Describe(ch)
	c.groupsCollector.Describe(ch)
	c.regionsCollector.Describe(ch)
	c.usersCollector.Describe(ch)
}

func (c *IdentityCollector) Collect(ch chan<- prometheus.Metric) {
	// Track if any sub-collector fails
	var hasError bool

	// Collect metrics from all sub-collectors
	if err := c.domainsCollector.Collect(ch); err != nil {
		hasError = true
	}
	if err := c.projectsCollector.Collect(ch); err != nil {
		hasError = true
	}
	if err := c.groupsCollector.Collect(ch); err != nil {
		hasError = true
	}
	if err := c.regionsCollector.Collect(ch); err != nil {
		hasError = true
	}
	if err := c.usersCollector.Collect(ch); err != nil {
		hasError = true
	}

	// Emit single up metric based on overall success/failure
	upValue := float64(1)
	if hasError {
		upValue = 0
	}
	ch <- prometheus.MustNewConstMetric(
		keystoneUpDesc,
		prometheus.GaugeValue,
		upValue,
	)
}
