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
					"id", "uuid", "display_name", "user_id", "project_id", "host",
					"availability_zone", "vm_state", "power_state", "task_state",
					"memory_mb", "vcpus", "root_gb", "ephemeral_gb",
					"launched_at", "terminated_at", "instance_type_id", "deleted",
				}).AddRow(
					1, "server-uuid-1", "test-server", "user-1", "project-1", "compute-1",
					"nova", "active", 1, nil,
					2048, 2, 20, 0,
					"2023-12-18 10:00:00", nil, 1, 0,
				).AddRow(
					2, "server-uuid-2", "test-server-2", "user-1", "project-1", "compute-2",
					"nova", "stopped", 4, nil,
					4096, 4, 40, 0,
					"2023-12-18 09:00:00", nil, 2, 0,
				)

				mock.ExpectQuery("SELECT (.+) FROM instances").WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "empty servers",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "uuid", "display_name", "user_id", "project_id", "host",
					"availability_zone", "vm_state", "power_state", "task_state",
					"memory_mb", "vcpus", "root_gb", "ephemeral_gb",
					"launched_at", "terminated_at", "instance_type_id", "deleted",
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
	_ = w.ServerCollector.Collect(ch)
}
