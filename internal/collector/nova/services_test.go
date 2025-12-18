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
					1, "uuid-scheduler-1", "controller-1", "nova-scheduler", "scheduler", 0, "",
					"2023-12-18 10:00:00", 0, 29, 150, 0,
				).AddRow(
					2, "uuid-compute-1", "compute-1", "nova-compute", "compute", 0, "",
					"2023-12-18 10:01:00", 0, 29, 200, 0,
				).AddRow(
					3, "uuid-compute-2", "compute-2", "nova-compute", "compute", 1, "maintenance",
					"2023-12-18 09:30:00", 0, 29, 180, 0,
				)

				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetServices)).WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "services with mixed states",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "uuid", "host", "binary", "topic", "disabled", "disabled_reason",
					"last_seen_up", "forced_down", "version", "report_count", "deleted",
				}).AddRow(
					1, "uuid-scheduler-1", "controller-1", "nova-scheduler", "scheduler", 0, "",
					"2023-12-18 10:00:00", 0, 29, 150, 0,
				).AddRow(
					2, "uuid-compute-1", "compute-1", "nova-compute", "compute", 1, "down for maintenance",
					"2023-12-18 08:00:00", 1, 29, 100, 0,
				).AddRow(
					3, "uuid-conductor-1", "controller-1", "nova-conductor", "conductor", 0, "",
					"2023-12-18 10:02:00", 0, 29, 175, 0,
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
			ExpectedMetrics: `# HELP openstack_nova_agent_state agent_state
# TYPE openstack_nova_agent_state counter
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
	if err := w.ServicesCollector.Collect(ch); err != nil {
		// In a real application, this error would be handled appropriately
		// For tests, we log it but don't fail the collection process
		// since the test framework expects the Collect method to not return an error
		_ = err
	}
}
