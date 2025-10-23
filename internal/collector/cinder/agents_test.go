package cinder

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	cinderdb "github.com/vexxhost/openstack_database_exporter/internal/db/cinder"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestAgentsCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with services",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"uuid", "host", "service", "admin_state",
					"zone", "disabled_reason", "state",
				}).AddRow(
					"3649e0f6-de80-ab6e-4f1c-351042d2f7fe", "devstack@lvmdriver-1", "cinder-volume", "enabled",
					"nova", nil, 1,
				).AddRow(
					"3649e0f6-de80-ab6e-4f1c-351042d2f7fe", "devstack", "cinder-scheduler", "enabled",
					"nova", "Test1", 1,
				).AddRow(
					"3649e0f6-de80-ab6e-4f1c-351042d2f7fe", "devstack", "cinder-backup", "enabled",
					"nova", "Test2", 1,
				)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetAllServices)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_cinder_agent_state agent_state
# TYPE openstack_cinder_agent_state gauge
openstack_cinder_agent_state{adminState="enabled",disabledReason="",hostname="devstack@lvmdriver-1",service="cinder-volume",uuid="3649e0f6-de80-ab6e-4f1c-351042d2f7fe",zone="nova"} 1
openstack_cinder_agent_state{adminState="enabled",disabledReason="Test1",hostname="devstack",service="cinder-scheduler",uuid="3649e0f6-de80-ab6e-4f1c-351042d2f7fe",zone="nova"} 1
openstack_cinder_agent_state{adminState="enabled",disabledReason="Test2",hostname="devstack",service="cinder-backup",uuid="3649e0f6-de80-ab6e-4f1c-351042d2f7fe",zone="nova"} 1
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetAllServices)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
			ExpectError:     true,
		},
	}

	testutil.RunCollectorTests(t, tests, NewAgentsCollector)
}
