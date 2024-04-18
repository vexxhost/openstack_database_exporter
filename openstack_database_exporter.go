package main

import (
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/vexxhost/openstack_database_exporter/collector"
)

var (
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	neutronDSN = kingpin.Flag(
		"neutron.dsn",
		"DSN for Neutron",
	).Envar("NEUTRON_DSN").Required().String()
	novaDSN = kingpin.Flag(
		"nova.dsn",
		"DSN for Nova",
	).Envar("NOVA_DSN").Required().String()
	octaviaDSN = kingpin.Flag(
		"octavia.dsn",
		"DSN for Octavia",
	).Envar("OCTAVIA_DSN").Required().String()
	toolkitFlags = webflag.AddFlags(kingpin.CommandLine, ":9180")
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)

	kingpin.Version(version.Print("openstack_database_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting openstack_database_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collector.NewNeutronDatabaseCollector(logger, *neutronDSN),
		collector.NewNovaDatabaseCollector(logger, *novaDSN),
		collector.NewOctaviaDatabaseCollector(logger, *octaviaDSN),
	)

	http.Handle(*metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	if *metricsPath != "/" && *metricsPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "OpenStack Exporter",
			Description: "Prometheus Exporter for OpenStack (via database)",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	srv := &http.Server{}
	if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
