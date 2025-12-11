package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/vexxhost/openstack_database_exporter/internal/collector/cinder"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/glance"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/keystone"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/magnum"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/manila"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/octavia"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/placement"
	"github.com/vexxhost/openstack_database_exporter/internal/dsn"
)

var (
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	toolkitFlags = webflag.AddFlags(kingpin.CommandLine, ":9180")

	// Database connection flags
	cinderDatabaseURL = kingpin.Flag(
		"cinder.database-url",
		"Cinder database connection URL (oslo.db format)",
	).Envar("CINDER_DATABASE_URL").String()
	glanceDatabaseURL = kingpin.Flag(
		"glance.database-url",
		"Glance database connection URL (oslo.db format)",
	).Envar("GLANCE_DATABASE_URL").String()
	keystoneDatabaseURL = kingpin.Flag(
		"keystone.database-url",
		"Keystone database connection URL (oslo.db format)",
	).Envar("KEYSTONE_DATABASE_URL").String()
	magnumDatabaseURL = kingpin.Flag(
		"magnum.database-url",
		"Magnum database connection URL (oslo.db format)",
	).Envar("MAGNUM_DATABASE_URL").String()
	manilaDatabaseURL = kingpin.Flag(
		"manila.database-url",
		"Manila database connection URL (oslo.db format)",
	).Envar("MANILA_DATABASE_URL").String()
	neutronDatabaseURL = kingpin.Flag(
		"neutron.database-url",
		"Neutron database connection URL (oslo.db format)",
	).Envar("NEUTRON_DATABASE_URL").String()
	octaviaDatabaseURL = kingpin.Flag(
		"octavia.database-url",
		"Octavia database connection URL (oslo.db format)",
	).Envar("OCTAVIA_DATABASE_URL").String()
	placementDatabaseURL = kingpin.Flag(
		"placement.database-url",
		"Placement database connection URL (oslo.db format)",
	).Envar("PLACEMENT_DATABASE_URL").String()
)

func main() {
	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)

	kingpin.Version(version.Print("openstack_database_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promslog.New(promslogConfig)

	logger.Info("Starting openstack_database_exporter", "version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())

	reg := prometheus.NewRegistry()
	registerCollectors(reg, logger)

	http.Handle(*metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	if *metricsPath != "/" && *metricsPath != "" {
		landingPage, err := web.NewLandingPage(web.LandingConfig{
			Name:        "OpenStack Database Exporter",
			Description: "Prometheus Exporter for OpenStack Databases",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{Address: *metricsPath, Text: "Metrics"},
			},
		})
		if err != nil {
			logger.Error("failed to create landing page", "err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	srv := &http.Server{}
	if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil {
		logger.Error("Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}

func registerCollectors(registry *prometheus.Registry, logger *slog.Logger) {
	// Cinder collectors
	if *cinderDatabaseURL != "" {
		db, err := connectDB(*cinderDatabaseURL, logger, "cinder")
		if err == nil {
			registry.MustRegister(cinder.NewAgentsCollector(db, logger))
			registry.MustRegister(cinder.NewLimitsCollector(db, logger))
			registry.MustRegister(cinder.NewPoolsCollector(db, logger))
			registry.MustRegister(cinder.NewSnapshotsCollector(db, logger))
			registry.MustRegister(cinder.NewVolumesCollector(db, logger))
			logger.Info("Registered Cinder collectors")
		}
	}

	// Glance collectors
	if *glanceDatabaseURL != "" {
		db, err := connectDB(*glanceDatabaseURL, logger, "glance")
		if err == nil {
			registry.MustRegister(glance.NewImagesCollector(db, logger))
			logger.Info("Registered Glance collectors")
		}
	}

	// Keystone collectors
	if *keystoneDatabaseURL != "" {
		db, err := connectDB(*keystoneDatabaseURL, logger, "keystone")
		if err == nil {
			registry.MustRegister(keystone.NewIdentityCollector(db, logger))
			logger.Info("Registered Keystone collectors")
		}
	}

	// Magnum collectors
	if *magnumDatabaseURL != "" {
		db, err := connectDB(*magnumDatabaseURL, logger, "magnum")
		if err == nil {
			registry.MustRegister(magnum.NewClustersCollector(db, logger))
			registry.MustRegister(magnum.NewMastersCollector(db, logger))
			registry.MustRegister(magnum.NewNodesCollector(db, logger))
			logger.Info("Registered Magnum collectors")
		}
	}

	// Manila collectors
	if *manilaDatabaseURL != "" {
		db, err := connectDB(*manilaDatabaseURL, logger, "manila")
		if err == nil {
			registry.MustRegister(manila.NewSharesCollector(db, logger))
			logger.Info("Registered Manila collectors")
		}
	}

	// Neutron collectors
	if *neutronDatabaseURL != "" {
		db, err := connectDB(*neutronDatabaseURL, logger, "neutron")
		if err == nil {
			registry.MustRegister(neutron.NewHARouterAgentPortBindingCollector(db, logger))
			logger.Info("Registered Neutron collectors")
		}
	}

	// Octavia collectors
	if *octaviaDatabaseURL != "" {
		db, err := connectDB(*octaviaDatabaseURL, logger, "octavia")
		if err == nil {
			registry.MustRegister(octavia.NewAmphoraCollector(db, logger))
			registry.MustRegister(octavia.NewLoadBalancerCollector(db, logger))
			registry.MustRegister(octavia.NewPoolCollector(db, logger))
			logger.Info("Registered Octavia collectors")
		}
	}

	// Placement collectors
	if *placementDatabaseURL != "" {
		db, err := connectDB(*placementDatabaseURL, logger, "placement")
		if err == nil {
			registry.MustRegister(placement.NewResourcesCollector(db, logger))
			logger.Info("Registered Placement collectors")
		}
	}
}

func connectDB(connectionString string, logger *slog.Logger, service string) (*sql.DB, error) {
	// Parse oslo.db-style connection string
	goDSN, err := dsn.ParseOsloDBConnectionString(connectionString)
	if err != nil {
		logger.Error("Failed to parse connection string",
			"service", service,
			"error", err,
		)
		return nil, err
	}

	db, err := sql.Open("mysql", goDSN)
	if err != nil {
		logger.Error("Failed to open database connection",
			"service", service,
			"error", err,
		)
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	// Test connection
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database",
			"service", service,
			"error", err,
		)
		db.Close()
		return nil, err
	}

	logger.Info("Connected to database", "service", service)
	return db, nil
}
