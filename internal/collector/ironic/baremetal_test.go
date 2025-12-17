package ironic

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	ironicdb "github.com/vexxhost/openstack_database_exporter/internal/db/ironic"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestBaremetalCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Mock the up check query
				rows := sqlmock.NewRows([]string{
					"uuid", "name", "power_state", "provision_state", "maintenance",
					"resource_class", "console_enabled", "retired", "retired_reason",
				}).AddRow(
					"550e8400-e29b-41d4-a716-446655440000", "node-1", "power on", "active", false,
					"baremetal", true, false, "",
				)
				mock.ExpectQuery(regexp.QuoteMeta(ironicdb.GetNodeMetrics)).WillReturnRows(rows)

				// Mock the nodes collector query
				rows2 := sqlmock.NewRows([]string{
					"uuid", "name", "power_state", "provision_state", "maintenance",
					"resource_class", "console_enabled", "retired", "retired_reason",
				}).AddRow(
					"550e8400-e29b-41d4-a716-446655440000", "node-1", "power on", "active", false,
					"baremetal", true, false, "",
				)
				mock.ExpectQuery(regexp.QuoteMeta(ironicdb.GetNodeMetrics)).WillReturnRows(rows2)
			},
			ExpectedMetrics: `# HELP openstack_ironic_node Ironic node status
# TYPE openstack_ironic_node gauge
openstack_ironic_node{console_enabled="true",id="550e8400-e29b-41d4-a716-446655440000",maintenance="false",name="node-1",power_state="power on",provision_state="active",resource_class="baremetal",retired="false",retired_reason=""} 1
# HELP openstack_ironic_up Whether the Ironic baremetal service is up
# TYPE openstack_ironic_up gauge
openstack_ironic_up 1
`,
		},
		{
			Name: "database connection failure",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Mock the up check query to fail
				mock.ExpectQuery(regexp.QuoteMeta(ironicdb.GetNodeMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_ironic_up Whether the Ironic baremetal service is up
# TYPE openstack_ironic_up gauge
openstack_ironic_up 0
`,
		},
		{
			Name: "nodes collector failure but service up",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Mock the up check query to succeed
				rows := sqlmock.NewRows([]string{
					"uuid", "name", "power_state", "provision_state", "maintenance",
					"resource_class", "console_enabled", "retired", "retired_reason",
				})
				mock.ExpectQuery(regexp.QuoteMeta(ironicdb.GetNodeMetrics)).WillReturnRows(rows)

				// Mock the nodes collector query to fail
				mock.ExpectQuery(regexp.QuoteMeta(ironicdb.GetNodeMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_ironic_up Whether the Ironic baremetal service is up
# TYPE openstack_ironic_up gauge
openstack_ironic_up 1
`,
		},
	}

	testutil.RunCollectorTests(t, tests, NewBaremetalCollector)
}
