package octavia

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	octaviadb "github.com/vexxhost/openstack_database_exporter/internal/db/octavia"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestPoolCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with pools",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "project_id", "name", "protocol", "lb_algorithm",
					"operating_status", "load_balancer_id", "provisioning_status",
				}).AddRow(
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
