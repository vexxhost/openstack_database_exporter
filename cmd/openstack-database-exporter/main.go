package main

import (
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/vexxhost/openstack_database_exporter/internal/collector"
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
	heatDatabaseURL = kingpin.Flag(
		"heat.database-url",
		"Heat database connection URL (oslo.db format)",
	).Envar("HEAT_DATABASE_URL").String()
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

	reg := collector.NewRegistry(collector.Config{
		CinderDatabaseURL:    *cinderDatabaseURL,
		GlanceDatabaseURL:    *glanceDatabaseURL,
		HeatDatabaseURL:      *heatDatabaseURL,
		KeystoneDatabaseURL:  *keystoneDatabaseURL,
		MagnumDatabaseURL:    *magnumDatabaseURL,
		ManilaDatabaseURL:    *manilaDatabaseURL,
		NeutronDatabaseURL:   *neutronDatabaseURL,
		OctaviaDatabaseURL:   *octaviaDatabaseURL,
		PlacementDatabaseURL: *placementDatabaseURL,
	}, logger)

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
