package collector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/vexxhost/openstack_database_exporter/internal/collector/cinder"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/glance"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/heat"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/keystone"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/magnum"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/manila"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/octavia"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/placement"
)

const (
	Namespace = "openstack"
)

type Config struct {
	CinderDatabaseURL    string
	GlanceDatabaseURL    string
	HeatDatabaseURL      string
	KeystoneDatabaseURL  string
	MagnumDatabaseURL    string
	ManilaDatabaseURL    string
	NeutronDatabaseURL   string
	OctaviaDatabaseURL   string
	PlacementDatabaseURL string
}

func NewRegistry(cfg Config, logger *slog.Logger) *prometheus.Registry {
	reg := prometheus.NewRegistry()

	cinder.RegisterCollectors(reg, cfg.CinderDatabaseURL, logger)
	glance.RegisterCollectors(reg, cfg.GlanceDatabaseURL, logger)
	heat.RegisterCollectors(reg, cfg.HeatDatabaseURL, logger)
	keystone.RegisterCollectors(reg, cfg.KeystoneDatabaseURL, logger)
	magnum.RegisterCollectors(reg, cfg.MagnumDatabaseURL, logger)
	manila.RegisterCollectors(reg, cfg.ManilaDatabaseURL, logger)
	neutron.RegisterCollectors(reg, cfg.NeutronDatabaseURL, logger)
	octavia.RegisterCollectors(reg, cfg.OctaviaDatabaseURL, logger)
	placement.RegisterCollectors(reg, cfg.PlacementDatabaseURL, logger)

	return reg
}
