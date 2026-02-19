package collector

import (
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/vexxhost/openstack_database_exporter/internal/collector/cinder"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/glance"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/heat"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/ironic"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/keystone"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/magnum"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/manila"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/nova"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/octavia"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/placement"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/project"
	"github.com/vexxhost/openstack_database_exporter/internal/db"
	keystonedb "github.com/vexxhost/openstack_database_exporter/internal/db/keystone"
)

const (
	Namespace = "openstack"
)

type Config struct {
	CinderDatabaseURL    string
	GlanceDatabaseURL    string
	HeatDatabaseURL      string
	IronicDatabaseURL    string
	KeystoneDatabaseURL  string
	MagnumDatabaseURL    string
	ManilaDatabaseURL    string
	NeutronDatabaseURL   string
	OctaviaDatabaseURL   string
	PlacementDatabaseURL string
	NovaDatabaseURL      string
	NovaAPIDatabaseURL   string
	ProjectCacheTTL      time.Duration
}

func NewRegistry(cfg Config, logger *slog.Logger) *prometheus.Registry {
	reg := prometheus.NewRegistry()

	// Create a single shared project resolver for all collectors that need
	// project ID → name resolution. This avoids duplicate keystone DB
	// connections and duplicate cache refreshes.
	var keystoneQueries *keystonedb.Queries
	if cfg.KeystoneDatabaseURL != "" {
		keystoneConn, err := db.Connect(cfg.KeystoneDatabaseURL)
		if err != nil {
			logger.Warn("Failed to connect to keystone database for project name resolution", "error", err)
		} else {
			keystoneQueries = keystonedb.New(keystoneConn)
		}
	}
	projectResolver := project.NewResolver(logger, keystoneQueries, cfg.ProjectCacheTTL)

	cinder.RegisterCollectors(reg, cfg.CinderDatabaseURL, projectResolver, logger)
	glance.RegisterCollectors(reg, cfg.GlanceDatabaseURL, logger)
	heat.RegisterCollectors(reg, cfg.HeatDatabaseURL, logger)
	ironic.RegisterCollectors(reg, cfg.IronicDatabaseURL, logger)
	keystone.RegisterCollectors(reg, cfg.KeystoneDatabaseURL, logger)
	magnum.RegisterCollectors(reg, cfg.MagnumDatabaseURL, logger)
	manila.RegisterCollectors(reg, cfg.ManilaDatabaseURL, logger)
	neutron.RegisterCollectors(reg, cfg.NeutronDatabaseURL, projectResolver, logger)
	nova.RegisterCollectors(reg, cfg.NovaDatabaseURL, cfg.NovaAPIDatabaseURL, cfg.PlacementDatabaseURL, projectResolver, logger)
	octavia.RegisterCollectors(reg, cfg.OctaviaDatabaseURL, logger)
	placement.RegisterCollectors(reg, cfg.PlacementDatabaseURL, logger)

	return reg
}
