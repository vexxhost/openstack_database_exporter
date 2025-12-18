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

func TestInstancesCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with instances data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "uuid", "display_name", "user_id", "project_id", "host", "availability_zone",
					"vm_state", "power_state", "task_state", "memory_mb", "vcpus", "root_gb", "ephemeral_gb",
					"launched_at", "terminated_at", "instance_type_id", "deleted",
				}).AddRow(
					1, "instance-1", "test-vm", "user-1", "project-1", "compute-1", "nova",
					"active", 1, nil, 4096, 2, 20, 0, "2023-01-01 12:00:00", nil, 1, 0,
				).AddRow(
					2, "instance-2", "test-vm-2", "user-1", "project-1", "compute-2", "nova",
					"stopped", 4, nil, 2048, 1, 10, 0, "2023-01-01 12:00:00", nil, 1, 0,
				)

				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetInstances)).WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "empty instances",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "uuid", "display_name", "user_id", "project_id", "host", "availability_zone",
					"vm_state", "power_state", "task_state", "memory_mb", "vcpus", "root_gb", "ephemeral_gb",
					"launched_at", "terminated_at", "instance_type_id", "deleted",
				})
				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetInstances)).WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetInstances)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		collector := NewInstancesCollector(logger, novadb.New(db), novaapidb.New(db))
		return &instancesCollectorWrapper{collector}
	})
}

type instancesCollectorWrapper struct {
	*InstancesCollector
}

func (w *instancesCollectorWrapper) Collect(ch chan<- prometheus.Metric) {
	w.InstancesCollector.Collect(ch)
}
