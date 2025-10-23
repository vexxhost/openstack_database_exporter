package octavia

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	octaviadb "github.com/vexxhost/openstack_database_exporter/internal/db/octavia"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestLoadBalancerCollector_Collect(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Mock load balancer query
				lbRows := sqlmock.NewRows([]string{
					"id", "project_id", "name", "provisioning_status",
					"operating_status", "provider", "vip_address",
				}).AddRow(
					"607226db-27ef-4d41-ae89-f2a800e9c2db", "e3cd678b11784734bc366148aa37580e", "best_load_balancer", "ACTIVE",
					"ONLINE", "octavia", "203.0.113.50",
				)

				mock.ExpectQuery(octaviadb.GetAllLoadBalancersWithVip).WillReturnRows(lbRows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_loadbalancer_status loadbalancer_status
# TYPE openstack_loadbalancer_loadbalancer_status gauge
openstack_loadbalancer_loadbalancer_status{id="607226db-27ef-4d41-ae89-f2a800e9c2db",name="best_load_balancer",operating_status="ONLINE",project_id="e3cd678b11784734bc366148aa37580e",provider="octavia",provisioning_status="ACTIVE",vip_address="203.0.113.50"} 0
# HELP openstack_loadbalancer_total_loadbalancers total_loadbalancers
# TYPE openstack_loadbalancer_total_loadbalancers gauge
openstack_loadbalancer_total_loadbalancers 1
# HELP openstack_loadbalancer_up up
# TYPE openstack_loadbalancer_up gauge
openstack_loadbalancer_up 1
`,
		},
		{
			Name: "handles query errors gracefully",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(octaviadb.GetAllLoadBalancersWithVip).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_up up
# TYPE openstack_loadbalancer_up gauge
openstack_loadbalancer_up 0
`,
		},
	}

	testutil.RunCollectorTests(t, tests, NewLoadBalancerCollector)
}
