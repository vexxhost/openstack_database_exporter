package placement

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	placementdb "github.com/vexxhost/openstack_database_exporter/internal/db/placement"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestResourcesCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with placement resources",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"hostname", "resource_type", "total", "allocation_ratio", "reserved", "used",
				}).AddRow(
					"cmp-1-svr8204.localdomain", "DISK_GB", 2047, "1.2000000476837158", 0, []uint8("6969"),
				).AddRow(
					"cmp-1-svr8204.localdomain", "MEMORY_MB", 772447, "1.2999999523162842", 8192, []uint8("1945"),
				).AddRow(
					"cmp-1-svr8204.localdomain", "VCPU", 96, "3.0000000000000000", 0, []uint8("10"),
				).AddRow(
					"cmp-5-svr8208.localdomain", "DISK_GB", 2047, "1.2000000476837158", 0, []uint8("0"),
				).AddRow(
					"cmp-5-svr8208.localdomain", "MEMORY_MB", 772447, "1.0000000000000000", 8192, []uint8("0"),
				).AddRow(
					"cmp-5-svr8208.localdomain", "PCPU", 96, "1.0000000000000000", 0, []uint8("0"),
				)

				mock.ExpectQuery(regexp.QuoteMeta(placementdb.GetResourceMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_placement_resource_allocation_ratio resource_allocation_ratio
# TYPE openstack_placement_resource_allocation_ratio gauge
openstack_placement_resource_allocation_ratio{hostname="cmp-1-svr8204.localdomain",resourcetype="DISK_GB"} 1.2000000476837158
openstack_placement_resource_allocation_ratio{hostname="cmp-1-svr8204.localdomain",resourcetype="MEMORY_MB"} 1.2999999523162842
openstack_placement_resource_allocation_ratio{hostname="cmp-1-svr8204.localdomain",resourcetype="VCPU"} 3
openstack_placement_resource_allocation_ratio{hostname="cmp-5-svr8208.localdomain",resourcetype="DISK_GB"} 1.2000000476837158
openstack_placement_resource_allocation_ratio{hostname="cmp-5-svr8208.localdomain",resourcetype="MEMORY_MB"} 1
openstack_placement_resource_allocation_ratio{hostname="cmp-5-svr8208.localdomain",resourcetype="PCPU"} 1
# HELP openstack_placement_resource_reserved resource_reserved
# TYPE openstack_placement_resource_reserved gauge
openstack_placement_resource_reserved{hostname="cmp-1-svr8204.localdomain",resourcetype="DISK_GB"} 0
openstack_placement_resource_reserved{hostname="cmp-1-svr8204.localdomain",resourcetype="MEMORY_MB"} 8192
openstack_placement_resource_reserved{hostname="cmp-1-svr8204.localdomain",resourcetype="VCPU"} 0
openstack_placement_resource_reserved{hostname="cmp-5-svr8208.localdomain",resourcetype="DISK_GB"} 0
openstack_placement_resource_reserved{hostname="cmp-5-svr8208.localdomain",resourcetype="MEMORY_MB"} 8192
openstack_placement_resource_reserved{hostname="cmp-5-svr8208.localdomain",resourcetype="PCPU"} 0
# HELP openstack_placement_resource_total resource_total
# TYPE openstack_placement_resource_total gauge
openstack_placement_resource_total{hostname="cmp-1-svr8204.localdomain",resourcetype="DISK_GB"} 2047
openstack_placement_resource_total{hostname="cmp-1-svr8204.localdomain",resourcetype="MEMORY_MB"} 772447
openstack_placement_resource_total{hostname="cmp-1-svr8204.localdomain",resourcetype="VCPU"} 96
openstack_placement_resource_total{hostname="cmp-5-svr8208.localdomain",resourcetype="DISK_GB"} 2047
openstack_placement_resource_total{hostname="cmp-5-svr8208.localdomain",resourcetype="MEMORY_MB"} 772447
openstack_placement_resource_total{hostname="cmp-5-svr8208.localdomain",resourcetype="PCPU"} 96
# HELP openstack_placement_resource_usage resource_usage
# TYPE openstack_placement_resource_usage gauge
openstack_placement_resource_usage{hostname="cmp-1-svr8204.localdomain",resourcetype="DISK_GB"} 6969
openstack_placement_resource_usage{hostname="cmp-1-svr8204.localdomain",resourcetype="MEMORY_MB"} 1945
openstack_placement_resource_usage{hostname="cmp-1-svr8204.localdomain",resourcetype="VCPU"} 10
openstack_placement_resource_usage{hostname="cmp-5-svr8208.localdomain",resourcetype="DISK_GB"} 0
openstack_placement_resource_usage{hostname="cmp-5-svr8208.localdomain",resourcetype="MEMORY_MB"} 0
openstack_placement_resource_usage{hostname="cmp-5-svr8208.localdomain",resourcetype="PCPU"} 0
# HELP openstack_placement_up up
# TYPE openstack_placement_up gauge
openstack_placement_up 1
`,
		},
		{
			Name: "successful collection with no resources",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"hostname", "resource_type", "total", "allocation_ratio", "reserved", "used",
				})

				mock.ExpectQuery(regexp.QuoteMeta(placementdb.GetResourceMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_placement_up up
# TYPE openstack_placement_up gauge
openstack_placement_up 1
`,
		},
		{
			Name: "handles null hostname gracefully",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"hostname", "resource_type", "total", "allocation_ratio", "reserved", "used",
				}).AddRow(
					sql.NullString{Valid: false}, "VCPU", 16, "2.0", 0, []uint8("4"),
				)

				mock.ExpectQuery(regexp.QuoteMeta(placementdb.GetResourceMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_placement_resource_allocation_ratio resource_allocation_ratio
# TYPE openstack_placement_resource_allocation_ratio gauge
openstack_placement_resource_allocation_ratio{hostname="",resourcetype="VCPU"} 2
# HELP openstack_placement_resource_reserved resource_reserved
# TYPE openstack_placement_resource_reserved gauge
openstack_placement_resource_reserved{hostname="",resourcetype="VCPU"} 0
# HELP openstack_placement_resource_total resource_total
# TYPE openstack_placement_resource_total gauge
openstack_placement_resource_total{hostname="",resourcetype="VCPU"} 16
# HELP openstack_placement_resource_usage resource_usage
# TYPE openstack_placement_resource_usage gauge
openstack_placement_resource_usage{hostname="",resourcetype="VCPU"} 4
# HELP openstack_placement_up up
# TYPE openstack_placement_up gauge
openstack_placement_up 1
`,
		},
		{
			Name: "handles invalid allocation ratio gracefully",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"hostname", "resource_type", "total", "allocation_ratio", "reserved", "used",
				}).AddRow(
					"test-host", "MEMORY_MB", 1024, "invalid_ratio", 0, []uint8("256"),
				)

				mock.ExpectQuery(regexp.QuoteMeta(placementdb.GetResourceMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_placement_resource_allocation_ratio resource_allocation_ratio
# TYPE openstack_placement_resource_allocation_ratio gauge
openstack_placement_resource_allocation_ratio{hostname="test-host",resourcetype="MEMORY_MB"} 1
# HELP openstack_placement_resource_reserved resource_reserved
# TYPE openstack_placement_resource_reserved gauge
openstack_placement_resource_reserved{hostname="test-host",resourcetype="MEMORY_MB"} 0
# HELP openstack_placement_resource_total resource_total
# TYPE openstack_placement_resource_total gauge
openstack_placement_resource_total{hostname="test-host",resourcetype="MEMORY_MB"} 1024
# HELP openstack_placement_resource_usage resource_usage
# TYPE openstack_placement_resource_usage gauge
openstack_placement_resource_usage{hostname="test-host",resourcetype="MEMORY_MB"} 256
# HELP openstack_placement_up up
# TYPE openstack_placement_up gauge
openstack_placement_up 1
`,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(placementdb.GetResourceMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_placement_up up
# TYPE openstack_placement_up gauge
openstack_placement_up 0
`,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) *ResourcesCollector {
		return NewResourcesCollector(db, logger)
	})
}