package heat

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	heatdb "github.com/vexxhost/openstack_database_exporter/internal/db/heat"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestStacksCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with stack data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "status", "action", "tenant",
				}).AddRow(
					"stack-1", "my-stack", "CREATE_COMPLETE", "CREATE", "tenant-1",
				).AddRow(
					"stack-2", "other-stack", "CREATE_COMPLETE", "CREATE", "tenant-2",
				).AddRow(
					"stack-3", "failed-stack", "CREATE_FAILED", "CREATE", "tenant-1",
				)

				mock.ExpectQuery(regexp.QuoteMeta(heatdb.GetStackMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_heat_stack_status_counter stack_status_counter
# TYPE openstack_heat_stack_status_counter gauge
openstack_heat_stack_status_counter{status="ADOPT_COMPLETE"} 0
openstack_heat_stack_status_counter{status="ADOPT_FAILED"} 0
openstack_heat_stack_status_counter{status="ADOPT_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="CHECK_COMPLETE"} 0
openstack_heat_stack_status_counter{status="CHECK_FAILED"} 0
openstack_heat_stack_status_counter{status="CHECK_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="CREATE_COMPLETE"} 2
openstack_heat_stack_status_counter{status="CREATE_FAILED"} 1
openstack_heat_stack_status_counter{status="CREATE_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="DELETE_COMPLETE"} 0
openstack_heat_stack_status_counter{status="DELETE_FAILED"} 0
openstack_heat_stack_status_counter{status="DELETE_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="INIT_COMPLETE"} 0
openstack_heat_stack_status_counter{status="INIT_FAILED"} 0
openstack_heat_stack_status_counter{status="INIT_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="RESUME_COMPLETE"} 0
openstack_heat_stack_status_counter{status="RESUME_FAILED"} 0
openstack_heat_stack_status_counter{status="RESUME_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="ROLLBACK_COMPLETE"} 0
openstack_heat_stack_status_counter{status="ROLLBACK_FAILED"} 0
openstack_heat_stack_status_counter{status="ROLLBACK_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="SNAPSHOT_COMPLETE"} 0
openstack_heat_stack_status_counter{status="SNAPSHOT_FAILED"} 0
openstack_heat_stack_status_counter{status="SNAPSHOT_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="SUSPEND_COMPLETE"} 0
openstack_heat_stack_status_counter{status="SUSPEND_FAILED"} 0
openstack_heat_stack_status_counter{status="SUSPEND_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="UPDATE_COMPLETE"} 0
openstack_heat_stack_status_counter{status="UPDATE_FAILED"} 0
openstack_heat_stack_status_counter{status="UPDATE_IN_PROGRESS"} 0
# HELP openstack_heat_up up
# TYPE openstack_heat_up gauge
openstack_heat_up 1
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "status", "action", "tenant",
				})

				mock.ExpectQuery(regexp.QuoteMeta(heatdb.GetStackMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_heat_stack_status_counter stack_status_counter
# TYPE openstack_heat_stack_status_counter gauge
openstack_heat_stack_status_counter{status="ADOPT_COMPLETE"} 0
openstack_heat_stack_status_counter{status="ADOPT_FAILED"} 0
openstack_heat_stack_status_counter{status="ADOPT_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="CHECK_COMPLETE"} 0
openstack_heat_stack_status_counter{status="CHECK_FAILED"} 0
openstack_heat_stack_status_counter{status="CHECK_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="CREATE_COMPLETE"} 0
openstack_heat_stack_status_counter{status="CREATE_FAILED"} 0
openstack_heat_stack_status_counter{status="CREATE_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="DELETE_COMPLETE"} 0
openstack_heat_stack_status_counter{status="DELETE_FAILED"} 0
openstack_heat_stack_status_counter{status="DELETE_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="INIT_COMPLETE"} 0
openstack_heat_stack_status_counter{status="INIT_FAILED"} 0
openstack_heat_stack_status_counter{status="INIT_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="RESUME_COMPLETE"} 0
openstack_heat_stack_status_counter{status="RESUME_FAILED"} 0
openstack_heat_stack_status_counter{status="RESUME_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="ROLLBACK_COMPLETE"} 0
openstack_heat_stack_status_counter{status="ROLLBACK_FAILED"} 0
openstack_heat_stack_status_counter{status="ROLLBACK_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="SNAPSHOT_COMPLETE"} 0
openstack_heat_stack_status_counter{status="SNAPSHOT_FAILED"} 0
openstack_heat_stack_status_counter{status="SNAPSHOT_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="SUSPEND_COMPLETE"} 0
openstack_heat_stack_status_counter{status="SUSPEND_FAILED"} 0
openstack_heat_stack_status_counter{status="SUSPEND_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="UPDATE_COMPLETE"} 0
openstack_heat_stack_status_counter{status="UPDATE_FAILED"} 0
openstack_heat_stack_status_counter{status="UPDATE_IN_PROGRESS"} 0
# HELP openstack_heat_up up
# TYPE openstack_heat_up gauge
openstack_heat_up 1
`,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(heatdb.GetStackMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_heat_up up
# TYPE openstack_heat_up gauge
openstack_heat_up 0
`,
		},
		{
			Name: "unknown status is ignored",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "status", "action", "tenant",
				}).AddRow(
					"stack-1", "my-stack", "CREATE_COMPLETE", "CREATE", "tenant-1",
				).AddRow(
					"stack-2", "odd-stack", "UNKNOWN_STATUS", "CREATE", "tenant-1",
				)

				mock.ExpectQuery(regexp.QuoteMeta(heatdb.GetStackMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_heat_stack_status_counter stack_status_counter
# TYPE openstack_heat_stack_status_counter gauge
openstack_heat_stack_status_counter{status="ADOPT_COMPLETE"} 0
openstack_heat_stack_status_counter{status="ADOPT_FAILED"} 0
openstack_heat_stack_status_counter{status="ADOPT_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="CHECK_COMPLETE"} 0
openstack_heat_stack_status_counter{status="CHECK_FAILED"} 0
openstack_heat_stack_status_counter{status="CHECK_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="CREATE_COMPLETE"} 1
openstack_heat_stack_status_counter{status="CREATE_FAILED"} 0
openstack_heat_stack_status_counter{status="CREATE_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="DELETE_COMPLETE"} 0
openstack_heat_stack_status_counter{status="DELETE_FAILED"} 0
openstack_heat_stack_status_counter{status="DELETE_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="INIT_COMPLETE"} 0
openstack_heat_stack_status_counter{status="INIT_FAILED"} 0
openstack_heat_stack_status_counter{status="INIT_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="RESUME_COMPLETE"} 0
openstack_heat_stack_status_counter{status="RESUME_FAILED"} 0
openstack_heat_stack_status_counter{status="RESUME_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="ROLLBACK_COMPLETE"} 0
openstack_heat_stack_status_counter{status="ROLLBACK_FAILED"} 0
openstack_heat_stack_status_counter{status="ROLLBACK_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="SNAPSHOT_COMPLETE"} 0
openstack_heat_stack_status_counter{status="SNAPSHOT_FAILED"} 0
openstack_heat_stack_status_counter{status="SNAPSHOT_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="SUSPEND_COMPLETE"} 0
openstack_heat_stack_status_counter{status="SUSPEND_FAILED"} 0
openstack_heat_stack_status_counter{status="SUSPEND_IN_PROGRESS"} 0
openstack_heat_stack_status_counter{status="UPDATE_COMPLETE"} 0
openstack_heat_stack_status_counter{status="UPDATE_FAILED"} 0
openstack_heat_stack_status_counter{status="UPDATE_IN_PROGRESS"} 0
# HELP openstack_heat_up up
# TYPE openstack_heat_up gauge
openstack_heat_up 1
`,
		},
	}

	testutil.RunCollectorTests(t, tests, NewStacksCollector)
}
