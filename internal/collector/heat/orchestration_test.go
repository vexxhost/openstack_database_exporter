package heat

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	heatdb "github.com/vexxhost/openstack_database_exporter/internal/db/heat"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestOrchestrationCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with stacks working",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "status", "action", "tenant",
				}).AddRow(
					"stack-1", "test-stack", "CREATE_COMPLETE", "CREATE", "project-1",
				)

				mock.ExpectQuery(regexp.QuoteMeta(heatdb.GetStackMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_orchestration_stack_status stack_status
# TYPE openstack_orchestration_stack_status gauge
openstack_orchestration_stack_status{action="CREATE",id="stack-1",name="test-stack",project_id="project-1",status="CREATE_COMPLETE"} 5
# HELP openstack_orchestration_stack_status_counter stack_status_counter
# TYPE openstack_orchestration_stack_status_counter gauge
openstack_orchestration_stack_status_counter{status="ADOPT_COMPLETE"} 0
openstack_orchestration_stack_status_counter{status="ADOPT_FAILED"} 0
openstack_orchestration_stack_status_counter{status="ADOPT_IN_PROGRESS"} 0
openstack_orchestration_stack_status_counter{status="CHECK_COMPLETE"} 0
openstack_orchestration_stack_status_counter{status="CHECK_FAILED"} 0
openstack_orchestration_stack_status_counter{status="CHECK_IN_PROGRESS"} 0
openstack_orchestration_stack_status_counter{status="CREATE_COMPLETE"} 1
openstack_orchestration_stack_status_counter{status="CREATE_FAILED"} 0
openstack_orchestration_stack_status_counter{status="CREATE_IN_PROGRESS"} 0
openstack_orchestration_stack_status_counter{status="DELETE_COMPLETE"} 0
openstack_orchestration_stack_status_counter{status="DELETE_FAILED"} 0
openstack_orchestration_stack_status_counter{status="DELETE_IN_PROGRESS"} 0
openstack_orchestration_stack_status_counter{status="INIT_COMPLETE"} 0
openstack_orchestration_stack_status_counter{status="INIT_FAILED"} 0
openstack_orchestration_stack_status_counter{status="INIT_IN_PROGRESS"} 0
openstack_orchestration_stack_status_counter{status="RESUME_COMPLETE"} 0
openstack_orchestration_stack_status_counter{status="RESUME_FAILED"} 0
openstack_orchestration_stack_status_counter{status="RESUME_IN_PROGRESS"} 0
openstack_orchestration_stack_status_counter{status="ROLLBACK_COMPLETE"} 0
openstack_orchestration_stack_status_counter{status="ROLLBACK_FAILED"} 0
openstack_orchestration_stack_status_counter{status="ROLLBACK_IN_PROGRESS"} 0
openstack_orchestration_stack_status_counter{status="SNAPSHOT_COMPLETE"} 0
openstack_orchestration_stack_status_counter{status="SNAPSHOT_FAILED"} 0
openstack_orchestration_stack_status_counter{status="SNAPSHOT_IN_PROGRESS"} 0
openstack_orchestration_stack_status_counter{status="SUSPEND_COMPLETE"} 0
openstack_orchestration_stack_status_counter{status="SUSPEND_FAILED"} 0
openstack_orchestration_stack_status_counter{status="SUSPEND_IN_PROGRESS"} 0
openstack_orchestration_stack_status_counter{status="UPDATE_COMPLETE"} 0
openstack_orchestration_stack_status_counter{status="UPDATE_FAILED"} 0
openstack_orchestration_stack_status_counter{status="UPDATE_IN_PROGRESS"} 0
# HELP openstack_orchestration_total_stacks total_stacks
# TYPE openstack_orchestration_total_stacks gauge
openstack_orchestration_total_stacks 1
# HELP openstack_orchestration_up up
# TYPE openstack_orchestration_up gauge
openstack_orchestration_up 1
`,
		},
		{
			Name: "stacks collector fails, up metric should be 0",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(heatdb.GetStackMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_orchestration_up up
# TYPE openstack_orchestration_up gauge
openstack_orchestration_up 0
`,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) *OrchestrationCollector {
		return NewOrchestrationCollector(db, logger)
	})
}
