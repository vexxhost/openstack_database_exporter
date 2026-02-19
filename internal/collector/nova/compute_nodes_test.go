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
					"id", "uuid", "host", "hypervisor_hostname", "hypervisor_type", "hypervisor_version",
					"vcpus", "vcpus_used", "memory_mb", "memory_mb_used", "local_gb", "local_gb_used",
					"disk_available_least", "free_ram_mb", "free_disk_gb", "current_workload",
					"running_vms", "cpu_allocation_ratio", "ram_allocation_ratio", "disk_allocation_ratio", "deleted",
				}).AddRow(
					1, "uuid-1", "compute-1", "compute-1.local", "QEMU", 4002001,
					16, 4, 32768, 8192, 1000, 200,
					800, 24576, 800, 2,
					3, 16.0, 1.5, 1.0, 0,
				).AddRow(
					2, "uuid-2", "compute-2", "compute-2.local", "QEMU", 4002001,
					32, 8, 65536, 16384, 2000, 400,
					1600, 49152, 1600, 4,
					6, 16.0, 1.5, 1.0, 0,
				)

				mock.ExpectQuery("SELECT (.+) FROM compute_nodes").WillReturnRows(rows)

				// Mock aggregate hosts query
				aggRows := sqlmock.NewRows([]string{
					"id", "host", "aggregate_id", "aggregate_name", "aggregate_uuid",
				}).AddRow(
					1, "compute-1.local", 1, "az1", "agg-uuid-1",
				)
				mock.ExpectQuery("SELECT (.+) FROM aggregate_hosts").WillReturnRows(aggRows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "empty compute nodes",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "uuid", "host", "hypervisor_hostname", "hypervisor_type", "hypervisor_version",
					"vcpus", "vcpus_used", "memory_mb", "memory_mb_used", "local_gb", "local_gb_used",
					"disk_available_least", "free_ram_mb", "free_disk_gb", "current_workload",
					"running_vms", "cpu_allocation_ratio", "ram_allocation_ratio", "disk_allocation_ratio", "deleted",
				})
				mock.ExpectQuery("SELECT (.+) FROM compute_nodes").WillReturnRows(rows)

				aggRows := sqlmock.NewRows([]string{
					"id", "host", "aggregate_id", "aggregate_name", "aggregate_uuid",
				})
				mock.ExpectQuery("SELECT (.+) FROM aggregate_hosts").WillReturnRows(aggRows)
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
