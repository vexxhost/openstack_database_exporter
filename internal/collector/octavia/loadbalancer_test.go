package octavia

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	octaviadb "github.com/vexxhost/openstack_database_exporter/internal/db/octavia"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestLoadBalancerCollector_Collect(t *testing.T) {
	cols := []string{
		"id", "project_id", "name", "provisioning_status",
		"operating_status", "provider", "vip_address",
	}

	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).AddRow(
					"607226db-27ef-4d41-ae89-f2a800e9c2db", "e3cd678b11784734bc366148aa37580e", "best_load_balancer", "ACTIVE",
					"ONLINE", "octavia", "203.0.113.50",
				)

				mock.ExpectQuery(octaviadb.GetAllLoadBalancersWithVip).WillReturnRows(rows)
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
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols)
				mock.ExpectQuery(octaviadb.GetAllLoadBalancersWithVip).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_total_loadbalancers total_loadbalancers
# TYPE openstack_loadbalancer_total_loadbalancers gauge
openstack_loadbalancer_total_loadbalancers 0
# HELP openstack_loadbalancer_up up
# TYPE openstack_loadbalancer_up gauge
openstack_loadbalancer_up 1
`,
		},
		{
			Name: "null optional fields",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).AddRow(
					"lb-001", nil, nil, "PENDING_CREATE",
					"OFFLINE", nil, nil,
				)
				mock.ExpectQuery(octaviadb.GetAllLoadBalancersWithVip).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_loadbalancer_status loadbalancer_status
# TYPE openstack_loadbalancer_loadbalancer_status gauge
openstack_loadbalancer_loadbalancer_status{id="lb-001",name="",operating_status="OFFLINE",project_id="",provider="",provisioning_status="PENDING_CREATE",vip_address=""} 2
# HELP openstack_loadbalancer_total_loadbalancers total_loadbalancers
# TYPE openstack_loadbalancer_total_loadbalancers gauge
openstack_loadbalancer_total_loadbalancers 1
# HELP openstack_loadbalancer_up up
# TYPE openstack_loadbalancer_up gauge
openstack_loadbalancer_up 1
`,
		},
		{
			Name: "all operating statuses",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).
					AddRow("lb-1", nil, nil, "ACTIVE", "ONLINE", nil, nil).
					AddRow("lb-2", nil, nil, "ACTIVE", "DRAINING", nil, nil).
					AddRow("lb-3", nil, nil, "ACTIVE", "OFFLINE", nil, nil).
					AddRow("lb-4", nil, nil, "ACTIVE", "ERROR", nil, nil).
					AddRow("lb-5", nil, nil, "ACTIVE", "NO_MONITOR", nil, nil).
					AddRow("lb-6", nil, nil, "ACTIVE", "UNKNOWN_OP", nil, nil)
				mock.ExpectQuery(octaviadb.GetAllLoadBalancersWithVip).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_loadbalancer_status loadbalancer_status
# TYPE openstack_loadbalancer_loadbalancer_status gauge
openstack_loadbalancer_loadbalancer_status{id="lb-1",name="",operating_status="ONLINE",project_id="",provider="",provisioning_status="ACTIVE",vip_address=""} 0
openstack_loadbalancer_loadbalancer_status{id="lb-2",name="",operating_status="DRAINING",project_id="",provider="",provisioning_status="ACTIVE",vip_address=""} 1
openstack_loadbalancer_loadbalancer_status{id="lb-3",name="",operating_status="OFFLINE",project_id="",provider="",provisioning_status="ACTIVE",vip_address=""} 2
openstack_loadbalancer_loadbalancer_status{id="lb-4",name="",operating_status="ERROR",project_id="",provider="",provisioning_status="ACTIVE",vip_address=""} 3
openstack_loadbalancer_loadbalancer_status{id="lb-5",name="",operating_status="NO_MONITOR",project_id="",provider="",provisioning_status="ACTIVE",vip_address=""} 4
openstack_loadbalancer_loadbalancer_status{id="lb-6",name="",operating_status="UNKNOWN_OP",project_id="",provider="",provisioning_status="ACTIVE",vip_address=""} -1
# HELP openstack_loadbalancer_total_loadbalancers total_loadbalancers
# TYPE openstack_loadbalancer_total_loadbalancers gauge
openstack_loadbalancer_total_loadbalancers 6
# HELP openstack_loadbalancer_up up
# TYPE openstack_loadbalancer_up gauge
openstack_loadbalancer_up 1
`,
		},
		{
			Name: "multiple load balancers with different providers",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).
					AddRow("lb-a", "proj-1", "web-lb", "ACTIVE", "ONLINE", "octavia", "10.0.0.1").
					AddRow("lb-b", "proj-2", "api-lb", "PENDING_UPDATE", "DRAINING", "amphora", "10.0.0.2").
					AddRow("lb-c", "proj-1", "internal-lb", "ERROR", "ERROR", "octavia", "10.0.0.3")
				mock.ExpectQuery(octaviadb.GetAllLoadBalancersWithVip).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_loadbalancer_status loadbalancer_status
# TYPE openstack_loadbalancer_loadbalancer_status gauge
openstack_loadbalancer_loadbalancer_status{id="lb-a",name="web-lb",operating_status="ONLINE",project_id="proj-1",provider="octavia",provisioning_status="ACTIVE",vip_address="10.0.0.1"} 0
openstack_loadbalancer_loadbalancer_status{id="lb-b",name="api-lb",operating_status="DRAINING",project_id="proj-2",provider="amphora",provisioning_status="PENDING_UPDATE",vip_address="10.0.0.2"} 1
openstack_loadbalancer_loadbalancer_status{id="lb-c",name="internal-lb",operating_status="ERROR",project_id="proj-1",provider="octavia",provisioning_status="ERROR",vip_address="10.0.0.3"} 3
# HELP openstack_loadbalancer_total_loadbalancers total_loadbalancers
# TYPE openstack_loadbalancer_total_loadbalancers gauge
openstack_loadbalancer_total_loadbalancers 3
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
