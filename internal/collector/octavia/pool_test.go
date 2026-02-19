package octavia

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	octaviadb "github.com/vexxhost/openstack_database_exporter/internal/db/octavia"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestPoolCollector(t *testing.T) {
	cols := []string{
		"id", "project_id", "name", "protocol", "lb_algorithm",
		"operating_status", "load_balancer_id", "provisioning_status",
	}

	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with pools",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).AddRow(
					"ca00ed86-94e3-440e-95c6-ffa35531081e", "8b1632d90bfe407787d9996b7f662fd7", "my_test_pool", "TCP", "ROUND_ROBIN",
					"ERROR", "e7284bb2-f46a-42ca-8c9b-e08671255125", "ACTIVE",
				)

				mock.ExpectQuery(octaviadb.GetAllPools).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_pool_status pool_status
# TYPE openstack_loadbalancer_pool_status gauge
openstack_loadbalancer_pool_status{id="ca00ed86-94e3-440e-95c6-ffa35531081e",lb_algorithm="ROUND_ROBIN",loadbalancers="e7284bb2-f46a-42ca-8c9b-e08671255125",name="my_test_pool",operating_status="ERROR",project_id="8b1632d90bfe407787d9996b7f662fd7",protocol="TCP",provisioning_status="ACTIVE"} 0
# HELP openstack_loadbalancer_total_pools total_pools
# TYPE openstack_loadbalancer_total_pools gauge
openstack_loadbalancer_total_pools 1
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols)
				mock.ExpectQuery(octaviadb.GetAllPools).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_total_pools total_pools
# TYPE openstack_loadbalancer_total_pools gauge
openstack_loadbalancer_total_pools 0
`,
		},
		{
			Name: "null optional fields",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).AddRow(
					"pool-001", nil, nil, "HTTP", "LEAST_CONNECTIONS",
					"ONLINE", nil, "PENDING_CREATE",
				)
				mock.ExpectQuery(octaviadb.GetAllPools).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_pool_status pool_status
# TYPE openstack_loadbalancer_pool_status gauge
openstack_loadbalancer_pool_status{id="pool-001",lb_algorithm="LEAST_CONNECTIONS",loadbalancers="",name="",operating_status="ONLINE",project_id="",protocol="HTTP",provisioning_status="PENDING_CREATE"} 3
# HELP openstack_loadbalancer_total_pools total_pools
# TYPE openstack_loadbalancer_total_pools gauge
openstack_loadbalancer_total_pools 1
`,
		},
		{
			Name: "all provisioning statuses",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).
					AddRow("p-1", nil, nil, "TCP", "ROUND_ROBIN", "ONLINE", nil, "ACTIVE").
					AddRow("p-2", nil, nil, "TCP", "ROUND_ROBIN", "ONLINE", nil, "DELETED").
					AddRow("p-3", nil, nil, "TCP", "ROUND_ROBIN", "ONLINE", nil, "ERROR").
					AddRow("p-4", nil, nil, "TCP", "ROUND_ROBIN", "ONLINE", nil, "PENDING_CREATE").
					AddRow("p-5", nil, nil, "TCP", "ROUND_ROBIN", "ONLINE", nil, "PENDING_UPDATE").
					AddRow("p-6", nil, nil, "TCP", "ROUND_ROBIN", "ONLINE", nil, "PENDING_DELETE").
					AddRow("p-7", nil, nil, "TCP", "ROUND_ROBIN", "ONLINE", nil, "UNKNOWN_PROV")
				mock.ExpectQuery(octaviadb.GetAllPools).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_pool_status pool_status
# TYPE openstack_loadbalancer_pool_status gauge
openstack_loadbalancer_pool_status{id="p-1",lb_algorithm="ROUND_ROBIN",loadbalancers="",name="",operating_status="ONLINE",project_id="",protocol="TCP",provisioning_status="ACTIVE"} 0
openstack_loadbalancer_pool_status{id="p-2",lb_algorithm="ROUND_ROBIN",loadbalancers="",name="",operating_status="ONLINE",project_id="",protocol="TCP",provisioning_status="DELETED"} 1
openstack_loadbalancer_pool_status{id="p-3",lb_algorithm="ROUND_ROBIN",loadbalancers="",name="",operating_status="ONLINE",project_id="",protocol="TCP",provisioning_status="ERROR"} 2
openstack_loadbalancer_pool_status{id="p-4",lb_algorithm="ROUND_ROBIN",loadbalancers="",name="",operating_status="ONLINE",project_id="",protocol="TCP",provisioning_status="PENDING_CREATE"} 3
openstack_loadbalancer_pool_status{id="p-5",lb_algorithm="ROUND_ROBIN",loadbalancers="",name="",operating_status="ONLINE",project_id="",protocol="TCP",provisioning_status="PENDING_UPDATE"} 4
openstack_loadbalancer_pool_status{id="p-6",lb_algorithm="ROUND_ROBIN",loadbalancers="",name="",operating_status="ONLINE",project_id="",protocol="TCP",provisioning_status="PENDING_DELETE"} 5
openstack_loadbalancer_pool_status{id="p-7",lb_algorithm="ROUND_ROBIN",loadbalancers="",name="",operating_status="ONLINE",project_id="",protocol="TCP",provisioning_status="UNKNOWN_PROV"} -1
# HELP openstack_loadbalancer_total_pools total_pools
# TYPE openstack_loadbalancer_total_pools gauge
openstack_loadbalancer_total_pools 7
`,
		},
		{
			Name: "multiple pools mixed protocols and algorithms",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).
					AddRow("pool-a", "proj-1", "http-pool", "HTTP", "ROUND_ROBIN", "ONLINE", "lb-1", "ACTIVE").
					AddRow("pool-b", "proj-1", "https-pool", "HTTPS", "LEAST_CONNECTIONS", "ERROR", "lb-1", "ACTIVE").
					AddRow("pool-c", "proj-2", "tcp-pool", "TCP", "SOURCE_IP", "OFFLINE", "lb-2", "PENDING_UPDATE")
				mock.ExpectQuery(octaviadb.GetAllPools).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_pool_status pool_status
# TYPE openstack_loadbalancer_pool_status gauge
openstack_loadbalancer_pool_status{id="pool-a",lb_algorithm="ROUND_ROBIN",loadbalancers="lb-1",name="http-pool",operating_status="ONLINE",project_id="proj-1",protocol="HTTP",provisioning_status="ACTIVE"} 0
openstack_loadbalancer_pool_status{id="pool-b",lb_algorithm="LEAST_CONNECTIONS",loadbalancers="lb-1",name="https-pool",operating_status="ERROR",project_id="proj-1",protocol="HTTPS",provisioning_status="ACTIVE"} 0
openstack_loadbalancer_pool_status{id="pool-c",lb_algorithm="SOURCE_IP",loadbalancers="lb-2",name="tcp-pool",operating_status="OFFLINE",project_id="proj-2",protocol="TCP",provisioning_status="PENDING_UPDATE"} 4
# HELP openstack_loadbalancer_total_pools total_pools
# TYPE openstack_loadbalancer_total_pools gauge
openstack_loadbalancer_total_pools 3
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(octaviadb.GetAllPools).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
			ExpectError:     true,
		},
	}

	testutil.RunCollectorTests(t, tests, NewPoolCollector)
}
