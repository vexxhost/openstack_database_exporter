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
					"pool1", "project1", "pool-one", "HTTP", "ROUND_ROBIN",
					"ONLINE", "lb1", "ACTIVE",
				).AddRow(
					"pool2", "project2", nil, "HTTPS", "LEAST_CONNECTIONS",
					"DEGRADED", "lb2", "ERROR",
				)

				mock.ExpectQuery(octaviadb.GetAllPools).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_pool_status pool_status
# TYPE openstack_loadbalancer_pool_status gauge
openstack_loadbalancer_pool_status{id="pool1",lb_algorithm="ROUND_ROBIN",loadbalancers="lb1",name="pool-one",operating_status="ONLINE",project_id="project1",protocol="HTTP",provisioning_status="ACTIVE"} 0
openstack_loadbalancer_pool_status{id="pool2",lb_algorithm="LEAST_CONNECTIONS",loadbalancers="lb2",name="",operating_status="DEGRADED",project_id="project2",protocol="HTTPS",provisioning_status="ERROR"} 2
# HELP openstack_loadbalancer_total_pools total_pools
# TYPE openstack_loadbalancer_total_pools gauge
openstack_loadbalancer_total_pools 2
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
