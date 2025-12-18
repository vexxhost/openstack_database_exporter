package nova

import (
	"database/sql"
	"log/slog"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	novadb "github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	novaapidb "github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestServerCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with server data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"uuid", "display_name", "user_id", "project_id", "power_state", "vm_state", "task_state",
				}).AddRow(
					"server-1", "test-server", "user-1", "project-1", 1, "active", nil,
				).AddRow(
					"server-2", "test-server-2", "user-1", "project-1", 4, "stopped", nil,
				)

				mock.ExpectQuery("SELECT (.+) FROM instances").WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "empty servers",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"uuid", "display_name", "user_id", "project_id", "power_state", "vm_state", "task_state",
				})
				mock.ExpectQuery("SELECT (.+) FROM instances").WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM instances").WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		collector := NewServerCollector(logger, novadb.New(db), novaapidb.New(db))
		return &serverCollectorWrapper{collector}
	})
}

type serverCollectorWrapper struct {
	*ServerCollector
}

func (w *serverCollectorWrapper) Collect(ch chan<- prometheus.Metric) {
	w.ServerCollector.Collect(ch)
}
