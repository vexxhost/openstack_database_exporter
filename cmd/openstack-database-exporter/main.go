package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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

const (
	defaultPort = "9180"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevel(),
	}))
	slog.SetDefault(logger)

	logger.Info("Starting OpenStack Database Exporter")

	// Create custom registry to avoid default Go metrics
	registry := prometheus.NewRegistry()

	// Register collectors based on configured database URLs
	registerCollectors(registry, logger)

	// Setup HTTP server
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OpenStack Database Exporter"))
	})

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	port := os.Getenv("EXPORTER_PORT")
	if port == "" {
		port = defaultPort
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("HTTP server listening", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", "error", err)
	}

	logger.Info("Shutdown complete")
}

func getLogLevel() slog.Level {
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug", "DEBUG":
		return slog.LevelDebug
	case "warn", "WARN", "warning", "WARNING":
		return slog.LevelWarn
	case "error", "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func registerCollectors(registry *prometheus.Registry, logger *slog.Logger) {
	// Cinder collectors
	if dbURL := os.Getenv("CINDER_DATABASE_URL"); dbURL != "" {
		db, err := connectDB(dbURL, logger, "cinder")
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
	if dbURL := os.Getenv("GLANCE_DATABASE_URL"); dbURL != "" {
		db, err := connectDB(dbURL, logger, "glance")
		if err == nil {
			registry.MustRegister(glance.NewImagesCollector(db, logger))
			logger.Info("Registered Glance collectors")
		}
	}

	// Keystone collectors
	if dbURL := os.Getenv("KEYSTONE_DATABASE_URL"); dbURL != "" {
		db, err := connectDB(dbURL, logger, "keystone")
		if err == nil {
			registry.MustRegister(keystone.NewIdentityCollector(db, logger))
			logger.Info("Registered Keystone collectors")
		}
	}

	// Magnum collectors
	if dbURL := os.Getenv("MAGNUM_DATABASE_URL"); dbURL != "" {
		db, err := connectDB(dbURL, logger, "magnum")
		if err == nil {
			registry.MustRegister(magnum.NewClustersCollector(db, logger))
			registry.MustRegister(magnum.NewMastersCollector(db, logger))
			registry.MustRegister(magnum.NewNodesCollector(db, logger))
			logger.Info("Registered Magnum collectors")
		}
	}

	// Manila collectors
	if dbURL := os.Getenv("MANILA_DATABASE_URL"); dbURL != "" {
		db, err := connectDB(dbURL, logger, "manila")
		if err == nil {
			registry.MustRegister(manila.NewSharesCollector(db, logger))
			logger.Info("Registered Manila collectors")
		}
	}

	// Neutron collectors
	if dbURL := os.Getenv("NEUTRON_DATABASE_URL"); dbURL != "" {
		db, err := connectDB(dbURL, logger, "neutron")
		if err == nil {
			registry.MustRegister(neutron.NewHARouterAgentPortBindingCollector(db, logger))
			logger.Info("Registered Neutron collectors")
		}
	}

	// Octavia collectors
	if dbURL := os.Getenv("OCTAVIA_DATABASE_URL"); dbURL != "" {
		db, err := connectDB(dbURL, logger, "octavia")
		if err == nil {
			registry.MustRegister(octavia.NewAmphoraCollector(db, logger))
			registry.MustRegister(octavia.NewLoadBalancerCollector(db, logger))
			registry.MustRegister(octavia.NewPoolCollector(db, logger))
			logger.Info("Registered Octavia collectors")
		}
	}

	// Placement collectors
	if dbURL := os.Getenv("PLACEMENT_DATABASE_URL"); dbURL != "" {
		db, err := connectDB(dbURL, logger, "placement")
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
	db.SetConnMaxLifetime(5 * time.Minute)

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
