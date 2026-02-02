package neutron

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/DATA-DOG/go-sqlmock"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestHARouterAgentPortBindingCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with active and standby routers",
			SetupMock: func(mock sqlmock.Sqlmock) {
				now := time.Now()
				rows := sqlmock.NewRows([]string{
					"router_id",
					"l3_agent_id",
					"state",
					"agent_host",
					"agent_admin_state_up",
					"agent_heartbeat_timestamp",
				}).AddRow(
					"9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f",
					"ddbf087c-e38f-4a73-bcb3-c38f2a719a03",
					"active",
					"dev-os-ctrl-02",
					1,
					now,
				).AddRow(
					"f8a44de0-fc8e-45df-93c7-f79bf3b01c95",
					"ddbf087c-e38f-4a73-bcb3-c38f2a719a03",
					"backup",
					"dev-os-ctrl-02",
					1,
					now,
				)

				mock.ExpectQuery(neutrondb.GetHARouterAgentPortBindingsWithAgents).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
# TYPE openstack_neutron_l3_agent_of_router gauge
openstack_neutron_l3_agent_of_router{agent_admin_up="true",agent_alive="true",agent_host="dev-os-ctrl-02",ha_state="active",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f"} 1
openstack_neutron_l3_agent_of_router{agent_admin_up="true",agent_alive="true",agent_host="dev-os-ctrl-02",ha_state="backup",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="f8a44de0-fc8e-45df-93c7-f79bf3b01c95"} 1
`,
		},
		{
			Name: "agent down when heartbeat too old",
			SetupMock: func(mock sqlmock.Sqlmock) {
				oldTime := time.Now().Add(-2 * time.Minute)
				rows := sqlmock.NewRows([]string{
					"router_id",
					"l3_agent_id",
					"state",
					"agent_host",
					"agent_admin_state_up",
					"agent_heartbeat_timestamp",
				}).AddRow(
					"9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f",
					"ddbf087c-e38f-4a73-bcb3-c38f2a719a03",
					"active",
					"dev-os-ctrl-02",
					1,
					oldTime,
				)

				mock.ExpectQuery(neutrondb.GetHARouterAgentPortBindingsWithAgents).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
# TYPE openstack_neutron_l3_agent_of_router gauge
openstack_neutron_l3_agent_of_router{agent_admin_up="true",agent_alive="false",agent_host="dev-os-ctrl-02",ha_state="active",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f"} 0
`,
		},
		{
			Name: "agent admin down",
			SetupMock: func(mock sqlmock.Sqlmock) {
				now := time.Now()
				rows := sqlmock.NewRows([]string{
					"router_id",
					"l3_agent_id",
					"state",
					"agent_host",
					"agent_admin_state_up",
					"agent_heartbeat_timestamp",
				}).AddRow(
					"9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f",
					"ddbf087c-e38f-4a73-bcb3-c38f2a719a03",
					"active",
					"dev-os-ctrl-02",
					0,
					now,
				)

				mock.ExpectQuery(neutrondb.GetHARouterAgentPortBindingsWithAgents).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
# TYPE openstack_neutron_l3_agent_of_router gauge
openstack_neutron_l3_agent_of_router{agent_admin_up="false",agent_alive="true",agent_host="dev-os-ctrl-02",ha_state="active",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f"} 1
`,
		},
		{
			Name: "handles NULL values from LEFT JOIN",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"router_id",
					"l3_agent_id",
					"state",
					"agent_host",
					"agent_admin_state_up",
					"agent_heartbeat_timestamp",
				}).AddRow(
					"9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f",
					"ddbf087c-e38f-4a73-bcb3-c38f2a719a03",
					nil,
					nil,
					nil,
					nil,
				)

				mock.ExpectQuery(neutrondb.GetHARouterAgentPortBindingsWithAgents).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
# TYPE openstack_neutron_l3_agent_of_router gauge
openstack_neutron_l3_agent_of_router{agent_admin_up="false",agent_alive="false",agent_host="",ha_state="",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f"} 0
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(neutrondb.GetHARouterAgentPortBindingsWithAgents).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
			ExpectError:     true,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		return &testHARouterAgentPortBindingCollector{NewHARouterAgentPortBindingCollector(db, logger)}
	})
}

type testHARouterAgentPortBindingCollector struct {
	*HARouterAgentPortBindingCollector
}

func (t *testHARouterAgentPortBindingCollector) Collect(ch chan<- prometheus.Metric) {
	_ = t.HARouterAgentPortBindingCollector.Collect(ch)
}

func TestRouterCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection of routers",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id",
					"name",
					"status",
					"admin_state_up",
					"project_id",
					"gw_port_id",
				}).AddRow(
					"f490f72e-f449-41be-857e-825096adacde",
					"router1",
					"ACTIVE",
					1,
					"d6fbbee0aa214c20b984292531ce7bd0",
					"547e89b2-f860-4aaf-b515-9a35b02f634d",
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetRouters)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_router router
# TYPE openstack_neutron_router gauge
openstack_neutron_router{admin_state_up="true",external_network_id="547e89b2-f860-4aaf-b515-9a35b02f634d",id="f490f72e-f449-41be-857e-825096adacde",name="router1",project_id="d6fbbee0aa214c20b984292531ce7bd0",status="ACTIVE"} 1
# HELP openstack_neutron_routers routers
# TYPE openstack_neutron_routers gauge
openstack_neutron_routers 1
# HELP openstack_neutron_routers_not_active routers_not_active
# TYPE openstack_neutron_routers_not_active gauge
openstack_neutron_routers_not_active 0
`,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		return &testRouterCollector{NewRouterCollector(db, logger)}
	})
}

type testRouterCollector struct {
	*RouterCollector
}

func (t *testRouterCollector) Collect(ch chan<- prometheus.Metric) {
	_ = t.RouterCollector.Collect(ch)
}
