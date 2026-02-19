package neutron

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestFloatingIPCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with floating IPs",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "floating_ip_address", "floating_network_id",
					"project_id", "router_id", "status", "fixed_ip_address",
				}).AddRow(
					"ce919300-9f7e-4f93-98e1-78236fb0f916",
					"10.13.55.227",
					"6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",
					"7a96a68dc8264f3d84fafd95a72265c5",
					nil,
					"DOWN",
					nil,
				).AddRow(
					"d0af13f7-c404-4dc7-8453-8f8b4d667b74",
					"10.13.55.238",
					"6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",
					"7a96a68dc8264f3d84fafd95a72265c5",
					"ede5fa94-ba7d-4902-8395-20feabb6146e",
					"ACTIVE",
					"10.16.0.113",
				)

				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetFloatingIPs)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_floating_ip floating_ip
# TYPE openstack_neutron_floating_ip gauge
openstack_neutron_floating_ip{floating_ip_address="10.13.55.227",floating_network_id="6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",id="ce919300-9f7e-4f93-98e1-78236fb0f916",project_id="7a96a68dc8264f3d84fafd95a72265c5",router_id="",status="DOWN"} 1
openstack_neutron_floating_ip{floating_ip_address="10.13.55.238",floating_network_id="6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",id="d0af13f7-c404-4dc7-8453-8f8b4d667b74",project_id="7a96a68dc8264f3d84fafd95a72265c5",router_id="ede5fa94-ba7d-4902-8395-20feabb6146e",status="ACTIVE"} 1
# HELP openstack_neutron_floating_ips floating_ips
# TYPE openstack_neutron_floating_ips gauge
openstack_neutron_floating_ips 2
# HELP openstack_neutron_floating_ips_associated_not_active floating_ips_associated_not_active
# TYPE openstack_neutron_floating_ips_associated_not_active gauge
openstack_neutron_floating_ips_associated_not_active 0
`,
		},
		{
			Name: "associated but not active floating IP",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "floating_ip_address", "floating_network_id",
					"project_id", "router_id", "status", "fixed_ip_address",
				}).AddRow(
					"fip-1", "10.0.0.1", "net-1", "proj-1",
					"router-1", "DOWN", "192.168.0.1",
				)

				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetFloatingIPs)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_floating_ip floating_ip
# TYPE openstack_neutron_floating_ip gauge
openstack_neutron_floating_ip{floating_ip_address="10.0.0.1",floating_network_id="net-1",id="fip-1",project_id="proj-1",router_id="router-1",status="DOWN"} 1
# HELP openstack_neutron_floating_ips floating_ips
# TYPE openstack_neutron_floating_ips gauge
openstack_neutron_floating_ips 1
# HELP openstack_neutron_floating_ips_associated_not_active floating_ips_associated_not_active
# TYPE openstack_neutron_floating_ips_associated_not_active gauge
openstack_neutron_floating_ips_associated_not_active 1
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "floating_ip_address", "floating_network_id",
					"project_id", "router_id", "status", "fixed_ip_address",
				})
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetFloatingIPs)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_floating_ips floating_ips
# TYPE openstack_neutron_floating_ips gauge
openstack_neutron_floating_ips 0
# HELP openstack_neutron_floating_ips_associated_not_active floating_ips_associated_not_active
# TYPE openstack_neutron_floating_ips_associated_not_active gauge
openstack_neutron_floating_ips_associated_not_active 0
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetFloatingIPs)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
		},
	}

	testutil.RunCollectorTests(t, tests, NewFloatingIPCollector)
}
