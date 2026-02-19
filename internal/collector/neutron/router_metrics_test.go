package neutron

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestRouterCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with routers",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "status", "admin_state_up",
					"project_id", "external_network_id",
				}).AddRow(
					"ede5fa94-ba7d-4902-8395-20feabb6146e",
					"private-router",
					"ACTIVE",
					true,
					"7a96a68dc8264f3d84fafd95a72265c5",
					"6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",
				)

				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetRouters)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_router router
# TYPE openstack_neutron_router gauge
openstack_neutron_router{admin_state_up="true",external_network_id="6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",id="ede5fa94-ba7d-4902-8395-20feabb6146e",name="private-router",project_id="7a96a68dc8264f3d84fafd95a72265c5",status="ACTIVE"} 1
# HELP openstack_neutron_routers routers
# TYPE openstack_neutron_routers gauge
openstack_neutron_routers 1
# HELP openstack_neutron_routers_not_active routers_not_active
# TYPE openstack_neutron_routers_not_active gauge
openstack_neutron_routers_not_active 0
`,
		},
		{
			Name: "router not active",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "status", "admin_state_up",
					"project_id", "external_network_id",
				}).AddRow(
					"router-1", "test-router", "ERROR", true,
					"proj-1", "",
				)

				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetRouters)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_router router
# TYPE openstack_neutron_router gauge
openstack_neutron_router{admin_state_up="true",external_network_id="",id="router-1",name="test-router",project_id="proj-1",status="ERROR"} 1
# HELP openstack_neutron_routers routers
# TYPE openstack_neutron_routers gauge
openstack_neutron_routers 1
# HELP openstack_neutron_routers_not_active routers_not_active
# TYPE openstack_neutron_routers_not_active gauge
openstack_neutron_routers_not_active 1
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "status", "admin_state_up",
					"project_id", "external_network_id",
				})
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetRouters)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_routers routers
# TYPE openstack_neutron_routers gauge
openstack_neutron_routers 0
# HELP openstack_neutron_routers_not_active routers_not_active
# TYPE openstack_neutron_routers_not_active gauge
openstack_neutron_routers_not_active 0
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetRouters)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
		},
	}

	testutil.RunCollectorTests(t, tests, NewRouterCollector)
}
