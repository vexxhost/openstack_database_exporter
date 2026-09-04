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
	placementdb "github.com/vexxhost/openstack_database_exporter/internal/db/placement"
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

				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetComputeNodes)).WillReturnRows(rows)

				// Mock aggregate host map: compute-1 in az1 (AZ-only agg) and s1 (non-AZ agg)
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetAggregateHostMap)).WillReturnRows(
					sqlmock.NewRows([]string{"host", "aggregate_id", "aggregate_name"}).
						AddRow("compute-1", 1, "az1").
						AddRow("compute-1", 2, "s1").
						AddRow("compute-2", 1, "az1").
						AddRow("compute-2", 3, "d1"),
				)

				// Mock aggregate metadata: agg 1 has only availability_zone=az1, agg 2 has cpu_allocation_ratio, agg 3 has something else
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetAggregateMetadata)).WillReturnRows(
					sqlmock.NewRows([]string{"aggregate_id", "key", "value"}).
						AddRow(1, "availability_zone", "az1").
						AddRow(2, "cpu_allocation_ratio", "16.0").
						AddRow(3, "custom_meta", "test"),
				)

				mock.ExpectQuery(regexp.QuoteMeta(placementdb.GetResourceMetrics)).WillReturnRows(
					sqlmock.NewRows([]string{"hostname", "resource_type", "total", "allocation_ratio", "reserved", "generation", "used"}),
				)
			},
			ExpectedMetrics: `# HELP openstack_nova_current_workload current_workload
# TYPE openstack_nova_current_workload gauge
openstack_nova_current_workload{aggregates="s1",availability_zone="az1",hostname="compute-1.local"} 2
openstack_nova_current_workload{aggregates="d1",availability_zone="az1",hostname="compute-2.local"} 4
# HELP openstack_nova_free_disk_bytes free_disk_bytes
# TYPE openstack_nova_free_disk_bytes gauge
openstack_nova_free_disk_bytes{aggregates="s1",availability_zone="az1",hostname="compute-1.local"} 8.589934592e+11
openstack_nova_free_disk_bytes{aggregates="d1",availability_zone="az1",hostname="compute-2.local"} 1.7179869184e+12
# HELP openstack_nova_local_storage_available_bytes local_storage_available_bytes
# TYPE openstack_nova_local_storage_available_bytes gauge
openstack_nova_local_storage_available_bytes{aggregates="s1",availability_zone="az1",hostname="compute-1.local"} 8.589934592e+11
openstack_nova_local_storage_available_bytes{aggregates="d1",availability_zone="az1",hostname="compute-2.local"} 1.7179869184e+12
# HELP openstack_nova_local_storage_used_bytes local_storage_used_bytes
# TYPE openstack_nova_local_storage_used_bytes gauge
openstack_nova_local_storage_used_bytes{aggregates="s1",availability_zone="az1",hostname="compute-1.local"} 2.147483648e+11
openstack_nova_local_storage_used_bytes{aggregates="d1",availability_zone="az1",hostname="compute-2.local"} 4.294967296e+11
# HELP openstack_nova_memory_available_bytes memory_available_bytes
# TYPE openstack_nova_memory_available_bytes gauge
openstack_nova_memory_available_bytes{aggregates="s1",availability_zone="az1",hostname="compute-1.local"} 2.5769803776e+10
openstack_nova_memory_available_bytes{aggregates="d1",availability_zone="az1",hostname="compute-2.local"} 5.1539607552e+10
# HELP openstack_nova_memory_used_bytes memory_used_bytes
# TYPE openstack_nova_memory_used_bytes gauge
openstack_nova_memory_used_bytes{aggregates="s1",availability_zone="az1",hostname="compute-1.local"} 8.589934592e+09
openstack_nova_memory_used_bytes{aggregates="d1",availability_zone="az1",hostname="compute-2.local"} 1.7179869184e+10
# HELP openstack_nova_running_vms running_vms
# TYPE openstack_nova_running_vms gauge
openstack_nova_running_vms{aggregates="s1",availability_zone="az1",hostname="compute-1.local"} 3
openstack_nova_running_vms{aggregates="d1",availability_zone="az1",hostname="compute-2.local"} 6
# HELP openstack_nova_vcpus_available vcpus_available
# TYPE openstack_nova_vcpus_available gauge
openstack_nova_vcpus_available{aggregates="s1",availability_zone="az1",hostname="compute-1.local"} 16
openstack_nova_vcpus_available{aggregates="d1",availability_zone="az1",hostname="compute-2.local"} 32
# HELP openstack_nova_vcpus_used vcpus_used
# TYPE openstack_nova_vcpus_used gauge
openstack_nova_vcpus_used{aggregates="s1",availability_zone="az1",hostname="compute-1.local"} 4
openstack_nova_vcpus_used{aggregates="d1",availability_zone="az1",hostname="compute-2.local"} 8
`,
		},
		{
			Name: "dedicated-cpu host falls back to PCPU total for vcpus_available",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// vcpus=0 because CPUs are tracked as PCPU inventory
				rows := sqlmock.NewRows([]string{
					"id", "uuid", "host", "hypervisor_hostname", "hypervisor_type", "hypervisor_version",
					"vcpus", "vcpus_used", "memory_mb", "memory_mb_used", "local_gb", "local_gb_used",
					"disk_available_least", "free_ram_mb", "free_disk_gb", "current_workload",
					"running_vms", "cpu_allocation_ratio", "ram_allocation_ratio", "disk_allocation_ratio", "deleted",
				}).AddRow(
					1, "uuid-1", "pinned-1", "pinned-1.local", "QEMU", 4002001,
					0, 16, 131072, 0, 1000, 0,
					800, 131072, 1000, 0,
					0, 1.0, 1.0, 1.0, 0,
				)

				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetComputeNodes)).WillReturnRows(rows)
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetAggregateHostMap)).WillReturnRows(
					sqlmock.NewRows([]string{"host", "aggregate_id", "aggregate_name"}).
						AddRow("pinned-1", 1, "g1"),
				)
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetAggregateMetadata)).WillReturnRows(
					sqlmock.NewRows([]string{"aggregate_id", "key", "value"}).
						AddRow(1, "custom_meta", "test"),
				)
				mock.ExpectQuery(regexp.QuoteMeta(placementdb.GetResourceMetrics)).WillReturnRows(
					sqlmock.NewRows([]string{"hostname", "resource_type", "total", "allocation_ratio", "reserved", "generation", "used"}).
						AddRow("pinned-1.local", "PCPU", 62, "1.0", 0, 5, []uint8("16")),
				)
			},
			ExpectedMetrics: `# HELP openstack_nova_current_workload current_workload
# TYPE openstack_nova_current_workload gauge
openstack_nova_current_workload{aggregates="g1",availability_zone="",hostname="pinned-1.local"} 0
# HELP openstack_nova_free_disk_bytes free_disk_bytes
# TYPE openstack_nova_free_disk_bytes gauge
openstack_nova_free_disk_bytes{aggregates="g1",availability_zone="",hostname="pinned-1.local"} 1.073741824e+12
# HELP openstack_nova_local_storage_available_bytes local_storage_available_bytes
# TYPE openstack_nova_local_storage_available_bytes gauge
openstack_nova_local_storage_available_bytes{aggregates="g1",availability_zone="",hostname="pinned-1.local"} 1.073741824e+12
# HELP openstack_nova_local_storage_used_bytes local_storage_used_bytes
# TYPE openstack_nova_local_storage_used_bytes gauge
openstack_nova_local_storage_used_bytes{aggregates="g1",availability_zone="",hostname="pinned-1.local"} 0
# HELP openstack_nova_memory_available_bytes memory_available_bytes
# TYPE openstack_nova_memory_available_bytes gauge
openstack_nova_memory_available_bytes{aggregates="g1",availability_zone="",hostname="pinned-1.local"} 1.37438953472e+11
# HELP openstack_nova_memory_used_bytes memory_used_bytes
# TYPE openstack_nova_memory_used_bytes gauge
openstack_nova_memory_used_bytes{aggregates="g1",availability_zone="",hostname="pinned-1.local"} 0
# HELP openstack_nova_running_vms running_vms
# TYPE openstack_nova_running_vms gauge
openstack_nova_running_vms{aggregates="g1",availability_zone="",hostname="pinned-1.local"} 0
# HELP openstack_nova_vcpus_available vcpus_available
# TYPE openstack_nova_vcpus_available gauge
openstack_nova_vcpus_available{aggregates="g1",availability_zone="",hostname="pinned-1.local"} 62
# HELP openstack_nova_vcpus_used vcpus_used
# TYPE openstack_nova_vcpus_used gauge
openstack_nova_vcpus_used{aggregates="g1",availability_zone="",hostname="pinned-1.local"} 16
`,
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
				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetComputeNodes)).WillReturnRows(rows)
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetAggregateHostMap)).WillReturnRows(
					sqlmock.NewRows([]string{"host", "aggregate_id", "aggregate_name"}),
				)
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetAggregateMetadata)).WillReturnRows(
					sqlmock.NewRows([]string{"aggregate_id", "key", "value"}),
				)
				mock.ExpectQuery(regexp.QuoteMeta(placementdb.GetResourceMetrics)).WillReturnRows(
					sqlmock.NewRows([]string{"hostname", "resource_type", "total", "allocation_ratio", "reserved", "generation", "used"}),
				)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetComputeNodes)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		collector := NewComputeNodesCollector(logger, novadb.New(db), novaapidb.New(db), placementdb.New(db))
		return &computeNodesCollectorWrapper{collector}
	})
}

type computeNodesCollectorWrapper struct {
	*ComputeNodesCollector
}

func (w *computeNodesCollectorWrapper) Collect(ch chan<- prometheus.Metric) {
	_ = w.ComputeNodesCollector.Collect(ch)
}
