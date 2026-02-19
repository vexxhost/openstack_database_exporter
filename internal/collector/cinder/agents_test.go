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
	cols := []string{
		"uuid", "host", "service", "admin_state",
		"zone", "disabled_reason", "state",
	}

	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with services",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).AddRow(
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
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetAllServices)).WillReturnRows(rows)
			},
			ExpectedMetrics: "",
		},
		{
			Name: "disabled agent with reason",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).AddRow(
					"aaaa-bbbb", "host-1", "cinder-volume", "disabled",
					"az-1", "maintenance window", 0,
				)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetAllServices)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_cinder_agent_state agent_state
# TYPE openstack_cinder_agent_state gauge
openstack_cinder_agent_state{adminState="disabled",disabledReason="maintenance window",hostname="host-1",service="cinder-volume",uuid="aaaa-bbbb",zone="az-1"} 0
`,
		},
		{
			Name: "null optional fields",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).AddRow(
					nil, nil, nil, "enabled",
					nil, nil, 1,
				)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetAllServices)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_cinder_agent_state agent_state
# TYPE openstack_cinder_agent_state gauge
openstack_cinder_agent_state{adminState="enabled",disabledReason="",hostname="",service="",uuid="",zone=""} 1
`,
		},
		{
			Name: "mixed enabled and disabled agents",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).
					AddRow("uuid-1", "host-a", "cinder-volume", "enabled", "nova", nil, 1).
					AddRow("uuid-2", "host-b", "cinder-scheduler", "disabled", "nova", "decommissioned", 0).
					AddRow("uuid-3", "host-c", "cinder-backup", "enabled", "az-2", nil, 0)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetAllServices)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_cinder_agent_state agent_state
# TYPE openstack_cinder_agent_state gauge
openstack_cinder_agent_state{adminState="enabled",disabledReason="",hostname="host-a",service="cinder-volume",uuid="uuid-1",zone="nova"} 1
openstack_cinder_agent_state{adminState="disabled",disabledReason="decommissioned",hostname="host-b",service="cinder-scheduler",uuid="uuid-2",zone="nova"} 0
openstack_cinder_agent_state{adminState="enabled",disabledReason="",hostname="host-c",service="cinder-backup",uuid="uuid-3",zone="az-2"} 0
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
