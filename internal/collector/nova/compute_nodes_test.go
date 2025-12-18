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

func TestComputeNodesCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with compute nodes data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"hypervisor_hostname", "vcpus", "memory_mb", "local_gb", "vcpus_used",
					"memory_mb_used", "local_gb_used", "running_vms", "hypervisor_type", "hypervisor_version",
				}).AddRow(
					"compute-1", 16, 32768, 1000, 4, 8192, 200, 2, "QEMU", 4002001,
				).AddRow(
					"compute-2", 32, 65536, 2000, 8, 16384, 400, 4, "QEMU", 4002001,
				)

				mock.ExpectQuery("SELECT (.+) FROM compute_nodes").WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "empty compute nodes",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"hypervisor_hostname", "vcpus", "memory_mb", "local_gb", "vcpus_used",
					"memory_mb_used", "local_gb_used", "running_vms", "hypervisor_type", "hypervisor_version",
				})
				mock.ExpectQuery("SELECT (.+) FROM compute_nodes").WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM compute_nodes").WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		collector := NewComputeNodesCollector(logger, novadb.New(db), novaapidb.New(db))
		return &computeNodesCollectorWrapper{collector}
	})
}

type computeNodesCollectorWrapper struct {
	*ComputeNodesCollector
}

func (w *computeNodesCollectorWrapper) Collect(ch chan<- prometheus.Metric) {
	_ = w.ComputeNodesCollector.Collect(ch)
}
