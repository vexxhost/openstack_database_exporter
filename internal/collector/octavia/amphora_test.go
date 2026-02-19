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
	cols := []string{
		"id", "compute_id", "status", "load_balancer_id",
		"lb_network_ip", "ha_ip", "role", "cert_expiration",
	}

	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with amphorae",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).AddRow(
					"45f40289-0551-483a-b089-47214bc2a8a4", "667bb225-69aa-44b1-8908-694dc624c267", "READY", "882f2a9d-9d53-4bd0-b0e9-08e9d0de11f9",
					"192.168.0.6", "10.0.0.6", "MASTER", time.Date(2020, 8, 8, 23, 44, 31, 0, time.UTC),
				).AddRow(
					"7f890893-ced0-46ed-8697-33415d070e5a", "9cd0f9a2-fe12-42fc-a7e3-5b6fbbe20395", "READY", "882f2a9d-9d53-4bd0-b0e9-08e9d0de11f9",
					"192.168.0.17", "10.0.0.6", "BACKUP", time.Date(2020, 8, 8, 23, 44, 30, 0, time.UTC),
				)

				mock.ExpectQuery(octaviadb.GetAllAmphora).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_amphora_status amphora_status
# TYPE openstack_loadbalancer_amphora_status gauge
openstack_loadbalancer_amphora_status{cert_expiration="2020-08-08T23:44:31Z",compute_id="667bb225-69aa-44b1-8908-694dc624c267",ha_ip="10.0.0.6",id="45f40289-0551-483a-b089-47214bc2a8a4",lb_network_ip="192.168.0.6",loadbalancer_id="882f2a9d-9d53-4bd0-b0e9-08e9d0de11f9",role="MASTER",status="READY"} 2
openstack_loadbalancer_amphora_status{cert_expiration="2020-08-08T23:44:30Z",compute_id="9cd0f9a2-fe12-42fc-a7e3-5b6fbbe20395",ha_ip="10.0.0.6",id="7f890893-ced0-46ed-8697-33415d070e5a",lb_network_ip="192.168.0.17",loadbalancer_id="882f2a9d-9d53-4bd0-b0e9-08e9d0de11f9",role="BACKUP",status="READY"} 2
# HELP openstack_loadbalancer_total_amphorae total_amphorae
# TYPE openstack_loadbalancer_total_amphorae gauge
openstack_loadbalancer_total_amphorae 2
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols)
				mock.ExpectQuery(octaviadb.GetAllAmphora).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_total_amphorae total_amphorae
# TYPE openstack_loadbalancer_total_amphorae gauge
openstack_loadbalancer_total_amphorae 0
`,
		},
		{
			Name: "null optional fields",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).AddRow(
					"aaa-bbb-ccc", nil, "BOOTING", nil,
					nil, nil, nil, nil,
				)
				mock.ExpectQuery(octaviadb.GetAllAmphora).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_amphora_status amphora_status
# TYPE openstack_loadbalancer_amphora_status gauge
openstack_loadbalancer_amphora_status{cert_expiration="",compute_id="",ha_ip="",id="aaa-bbb-ccc",lb_network_ip="",loadbalancer_id="",role="",status="BOOTING"} 0
# HELP openstack_loadbalancer_total_amphorae total_amphorae
# TYPE openstack_loadbalancer_total_amphorae gauge
openstack_loadbalancer_total_amphorae 1
`,
		},
		{
			Name: "all amphora statuses",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).
					AddRow("id-1", nil, "BOOTING", nil, nil, nil, nil, nil).
					AddRow("id-2", nil, "ALLOCATED", nil, nil, nil, nil, nil).
					AddRow("id-3", nil, "READY", nil, nil, nil, nil, nil).
					AddRow("id-4", nil, "PENDING_CREATE", nil, nil, nil, nil, nil).
					AddRow("id-5", nil, "PENDING_DELETE", nil, nil, nil, nil, nil).
					AddRow("id-6", nil, "DELETED", nil, nil, nil, nil, nil).
					AddRow("id-7", nil, "ERROR", nil, nil, nil, nil, nil).
					AddRow("id-8", nil, "UNKNOWN_STATUS", nil, nil, nil, nil, nil)
				mock.ExpectQuery(octaviadb.GetAllAmphora).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_amphora_status amphora_status
# TYPE openstack_loadbalancer_amphora_status gauge
openstack_loadbalancer_amphora_status{cert_expiration="",compute_id="",ha_ip="",id="id-1",lb_network_ip="",loadbalancer_id="",role="",status="BOOTING"} 0
openstack_loadbalancer_amphora_status{cert_expiration="",compute_id="",ha_ip="",id="id-2",lb_network_ip="",loadbalancer_id="",role="",status="ALLOCATED"} 1
openstack_loadbalancer_amphora_status{cert_expiration="",compute_id="",ha_ip="",id="id-3",lb_network_ip="",loadbalancer_id="",role="",status="READY"} 2
openstack_loadbalancer_amphora_status{cert_expiration="",compute_id="",ha_ip="",id="id-4",lb_network_ip="",loadbalancer_id="",role="",status="PENDING_CREATE"} 3
openstack_loadbalancer_amphora_status{cert_expiration="",compute_id="",ha_ip="",id="id-5",lb_network_ip="",loadbalancer_id="",role="",status="PENDING_DELETE"} 4
openstack_loadbalancer_amphora_status{cert_expiration="",compute_id="",ha_ip="",id="id-6",lb_network_ip="",loadbalancer_id="",role="",status="DELETED"} 5
openstack_loadbalancer_amphora_status{cert_expiration="",compute_id="",ha_ip="",id="id-7",lb_network_ip="",loadbalancer_id="",role="",status="ERROR"} 6
openstack_loadbalancer_amphora_status{cert_expiration="",compute_id="",ha_ip="",id="id-8",lb_network_ip="",loadbalancer_id="",role="",status="UNKNOWN_STATUS"} -1
# HELP openstack_loadbalancer_total_amphorae total_amphorae
# TYPE openstack_loadbalancer_total_amphorae gauge
openstack_loadbalancer_total_amphorae 8
`,
		},
		{
			Name: "single amphora with cert expiration",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).AddRow(
					"single-id", "compute-123", "ALLOCATED", "lb-456",
					"10.0.0.1", "10.0.0.2", "STANDALONE", time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
				)
				mock.ExpectQuery(octaviadb.GetAllAmphora).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_loadbalancer_amphora_status amphora_status
# TYPE openstack_loadbalancer_amphora_status gauge
openstack_loadbalancer_amphora_status{cert_expiration="2025-12-31T23:59:59Z",compute_id="compute-123",ha_ip="10.0.0.2",id="single-id",lb_network_ip="10.0.0.1",loadbalancer_id="lb-456",role="STANDALONE",status="ALLOCATED"} 1
# HELP openstack_loadbalancer_total_amphorae total_amphorae
# TYPE openstack_loadbalancer_total_amphorae gauge
openstack_loadbalancer_total_amphorae 1
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
