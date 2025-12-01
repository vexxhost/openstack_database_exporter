package keystone

import (
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
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
	c.domainsCollector.Describe(ch)
	c.projectsCollector.Describe(ch)
	c.groupsCollector.Describe(ch)
	c.regionsCollector.Describe(ch)
	c.usersCollector.Describe(ch)
}

func (c *IdentityCollector) Collect(ch chan<- prometheus.Metric) {
	// Collect metrics from all sub-collectors
	c.domainsCollector.Collect(ch)
	c.projectsCollector.Collect(ch)
	c.groupsCollector.Collect(ch)
	c.regionsCollector.Collect(ch)
	c.usersCollector.Collect(ch)
}
