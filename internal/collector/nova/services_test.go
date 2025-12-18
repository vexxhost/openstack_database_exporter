package nova

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	novadb "github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	novaapidb "github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestServicesCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with services data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "uuid", "host", "binary", "topic", "disabled", "disabled_reason",
					"last_seen_up", "forced_down", "version", "report_count", "deleted",
				}).AddRow(
					1, "uuid-1", "host1", "nova-scheduler", "scheduler", 1, "test1",
					"2023-01-01 12:00:00", 0, 1, 10, 0,
				).AddRow(
					2, "uuid-2", "host1", "nova-compute", "compute", 1, "test2",
					"2023-01-01 12:00:00", 0, 1, 10, 0,
				)

				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetServices)).WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "empty services",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "uuid", "host", "binary", "topic", "disabled", "disabled_reason",
					"last_seen_up", "forced_down", "version", "report_count", "deleted",
				})
				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetServices)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_nova_services services
# TYPE openstack_nova_services gauge
openstack_nova_services 0
`,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetServices)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		// Create a wrapper that implements prometheus.Collector properly
		collector := NewServicesCollector(logger, novadb.New(db), novaapidb.New(db))
		return &servicesCollectorWrapper{collector}
	})
}

// Wrapper to adapt ServicesCollector to prometheus.Collector interface
type servicesCollectorWrapper struct {
	*ServicesCollector
}

func (w *servicesCollectorWrapper) Collect(ch chan<- prometheus.Metric) {
	w.ServicesCollector.Collect(ch) // Ignoring error for test simplicity
}
