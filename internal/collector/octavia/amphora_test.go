package octavia

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	octaviadb "github.com/vexxhost/openstack_database_exporter/internal/db/octavia"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestAmphoraCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with amphorae",
			SetupMock: func(mock sqlmock.Sqlmock) {
				certTime := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
				rows := sqlmock.NewRows([]string{
					"id", "compute_id", "status", "load_balancer_id",
					"lb_network_ip", "ha_ip", "role", "cert_expiration",
				}).AddRow(
					"amp1", "compute1", "ALLOCATED", "lb1",
					"10.0.0.5", "192.168.1.10", "MASTER", certTime,
				).AddRow(
					"amp2", "compute2", "ERROR", nil,
					nil, nil, nil, nil,
				)

				mock.ExpectQuery(octaviadb.GetAllAmphora).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_amphora_status amphora_status
# TYPE openstack_loadbalancer_amphora_status gauge
openstack_loadbalancer_amphora_status{cert_expiration="2024-12-31T23:59:59Z",compute_id="compute1",ha_ip="192.168.1.10",id="amp1",lb_network_ip="10.0.0.5",loadbalancer_id="lb1",role="MASTER",status="ALLOCATED"} 1
openstack_loadbalancer_amphora_status{cert_expiration="",compute_id="compute2",ha_ip="",id="amp2",lb_network_ip="",loadbalancer_id="",role="",status="ERROR"} 6
# HELP openstack_loadbalancer_total_amphorae total_amphorae
# TYPE openstack_loadbalancer_total_amphorae gauge
openstack_loadbalancer_total_amphorae 2
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(octaviadb.GetAllAmphora).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
			ExpectError:     true,
		},
	}

	testutil.RunCollectorTests(t, tests, NewAmphoraCollector)
}
