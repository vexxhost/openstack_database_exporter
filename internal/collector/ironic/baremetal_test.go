package ironic

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	ironicdb "github.com/vexxhost/openstack_database_exporter/internal/db/ironic"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestBaremetalCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with single node",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"uuid", "name", "power_state", "provision_state", "maintenance",
					"resource_class", "console_enabled", "retired", "retired_reason",
				}).AddRow(
					"550e8400-e29b-41d4-a716-446655440000", "node-1", "power on", "active", false,
					"baremetal", true, false, "",
				)
				mock.ExpectQuery(regexp.QuoteMeta(ironicdb.GetNodeMetrics)).WillReturnRows(rows)
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
			Name: "multiple nodes with varied states",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"uuid", "name", "power_state", "provision_state", "maintenance",
					"resource_class", "console_enabled", "retired", "retired_reason",
				}).AddRow(
					"aaa-bbb-ccc", "node-active", "power on", "active", false,
					"baremetal", false, false, "",
				).AddRow(
					"ddd-eee-fff", "node-maint", "power on", "active", true,
					"baremetal", false, false, "",
				).AddRow(
					"ggg-hhh-iii", "node-retired", "power off", "manageable", false,
					"baremetal", false, true, "end of life",
				)
				mock.ExpectQuery(regexp.QuoteMeta(ironicdb.GetNodeMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_ironic_node Ironic node status
# TYPE openstack_ironic_node gauge
openstack_ironic_node{console_enabled="false",id="aaa-bbb-ccc",maintenance="false",name="node-active",power_state="power on",provision_state="active",resource_class="baremetal",retired="false",retired_reason=""} 1
openstack_ironic_node{console_enabled="false",id="ddd-eee-fff",maintenance="true",name="node-maint",power_state="power on",provision_state="active",resource_class="baremetal",retired="false",retired_reason=""} 1
openstack_ironic_node{console_enabled="false",id="ggg-hhh-iii",maintenance="false",name="node-retired",power_state="power off",provision_state="manageable",resource_class="baremetal",retired="true",retired_reason="end of life"} 1
# HELP openstack_ironic_up Whether the Ironic baremetal service is up
# TYPE openstack_ironic_up gauge
openstack_ironic_up 1
`,
		},
		{
			Name: "database connection failure",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(ironicdb.GetNodeMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_ironic_up Whether the Ironic baremetal service is up
# TYPE openstack_ironic_up gauge
openstack_ironic_up 0
`,
		},
		{
			Name: "skips nodes with null uuid",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"uuid", "name", "power_state", "provision_state", "maintenance",
					"resource_class", "console_enabled", "retired", "retired_reason",
				}).AddRow(
					nil, "node-no-uuid", "power on", "active", false,
					"baremetal", true, false, "",
				).AddRow(
					"550e8400-e29b-41d4-a716-446655440000", "node-1", "power on", "active", false,
					"baremetal", true, false, "",
				)
				mock.ExpectQuery(regexp.QuoteMeta(ironicdb.GetNodeMetrics)).WillReturnRows(rows)
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
			Name: "empty result set",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"uuid", "name", "power_state", "provision_state", "maintenance",
					"resource_class", "console_enabled", "retired", "retired_reason",
				})
				mock.ExpectQuery(regexp.QuoteMeta(ironicdb.GetNodeMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_ironic_up Whether the Ironic baremetal service is up
# TYPE openstack_ironic_up gauge
openstack_ironic_up 1
`,
		},
		{
			Name: "node with null optional fields",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"uuid", "name", "power_state", "provision_state", "maintenance",
					"resource_class", "console_enabled", "retired", "retired_reason",
				}).AddRow(
					"uuid-123", nil, nil, nil, nil,
					nil, nil, nil, "",
				)
				mock.ExpectQuery(regexp.QuoteMeta(ironicdb.GetNodeMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_ironic_node Ironic node status
# TYPE openstack_ironic_node gauge
openstack_ironic_node{console_enabled="false",id="uuid-123",maintenance="false",name="",power_state="unknown",provision_state="unknown",resource_class="unknown",retired="false",retired_reason=""} 1
# HELP openstack_ironic_up Whether the Ironic baremetal service is up
# TYPE openstack_ironic_up gauge
openstack_ironic_up 1
`,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		return NewBaremetalCollector(db, logger)
	})
}
