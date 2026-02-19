package neutron

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestSubnetCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with subnets and IP availability",
			SetupMock: func(mock sqlmock.Sqlmock) {
				subnetRows := sqlmock.NewRows([]string{
					"id", "name", "cidr", "gateway_ip", "network_id",
					"project_id", "enable_dhcp", "dns_nameservers",
					"subnetpool_id", "tags",
				}).AddRow(
					"5b32ccf9-ddbe-402b-9b68-bc66cf3c20ce",
					"public-subnet",
					"10.13.55.0/24",
					"10.13.55.1",
					"6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",
					"da457edfad314ed98fc84ef5e7d37f37",
					false,
					[]byte(""),
					nil,
					[]byte(""),
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnets)).WillReturnRows(subnetRows)

				usedRows := sqlmock.NewRows([]string{
					"subnet_id", "subnet_name", "cidr", "ip_version",
					"project_id", "network_id", "network_name", "allocation_count",
				}).AddRow(
					"5b32ccf9-ddbe-402b-9b68-bc66cf3c20ce",
					"public-subnet",
					"10.13.55.0/24",
					4,
					"da457edfad314ed98fc84ef5e7d37f37",
					"6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",
					"public",
					3,
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetNetworkIPAvailabilitiesUsed)).WillReturnRows(usedRows)

				totalRows := sqlmock.NewRows([]string{
					"subnet_name", "network_name", "subnet_id", "network_id",
					"first_ip", "last_ip", "project_id", "cidr", "ip_version",
				}).AddRow(
					"public-subnet",
					"public",
					"5b32ccf9-ddbe-402b-9b68-bc66cf3c20ce",
					"6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",
					"10.13.55.230",
					"10.13.55.249",
					"da457edfad314ed98fc84ef5e7d37f37",
					"10.13.55.0/24",
					4,
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetNetworkIPAvailabilitiesTotal)).WillReturnRows(totalRows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_network_ip_availabilities_total network_ip_availabilities_total
# TYPE openstack_neutron_network_ip_availabilities_total gauge
openstack_neutron_network_ip_availabilities_total{cidr="10.13.55.0/24",ip_version="4",network_id="6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",network_name="public",project_id="da457edfad314ed98fc84ef5e7d37f37",subnet_name="public-subnet"} 20
# HELP openstack_neutron_network_ip_availabilities_used network_ip_availabilities_used
# TYPE openstack_neutron_network_ip_availabilities_used gauge
openstack_neutron_network_ip_availabilities_used{cidr="10.13.55.0/24",ip_version="4",network_id="6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",network_name="public",project_id="da457edfad314ed98fc84ef5e7d37f37",subnet_name="public-subnet"} 3
# HELP openstack_neutron_subnet subnet
# TYPE openstack_neutron_subnet gauge
openstack_neutron_subnet{cidr="10.13.55.0/24",dns_nameservers="",enable_dhcp="false",gateway_ip="10.13.55.1",id="5b32ccf9-ddbe-402b-9b68-bc66cf3c20ce",name="public-subnet",network_id="6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",tags="",tenant_id="da457edfad314ed98fc84ef5e7d37f37"} 1
# HELP openstack_neutron_subnets subnets
# TYPE openstack_neutron_subnets gauge
openstack_neutron_subnets 1
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnets)).WillReturnRows(
					sqlmock.NewRows([]string{
						"id", "name", "cidr", "gateway_ip", "network_id",
						"project_id", "enable_dhcp", "dns_nameservers",
						"subnetpool_id", "tags",
					}))
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetNetworkIPAvailabilitiesUsed)).WillReturnRows(
					sqlmock.NewRows([]string{
						"subnet_id", "subnet_name", "cidr", "ip_version",
						"project_id", "network_id", "network_name", "allocation_count",
					}))
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetNetworkIPAvailabilitiesTotal)).WillReturnRows(
					sqlmock.NewRows([]string{
						"subnet_name", "network_name", "subnet_id", "network_id",
						"first_ip", "last_ip", "project_id", "cidr", "ip_version",
					}))
			},
			ExpectedMetrics: `# HELP openstack_neutron_subnets subnets
# TYPE openstack_neutron_subnets gauge
openstack_neutron_subnets 0
`,
		},
		{
			Name: "subnet query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnets)).WillReturnError(sql.ErrConnDone)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetNetworkIPAvailabilitiesUsed)).WillReturnRows(
					sqlmock.NewRows([]string{
						"subnet_id", "subnet_name", "cidr", "ip_version",
						"project_id", "network_id", "network_name", "allocation_count",
					}))
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetNetworkIPAvailabilitiesTotal)).WillReturnRows(
					sqlmock.NewRows([]string{
						"subnet_name", "network_name", "subnet_id", "network_id",
						"first_ip", "last_ip", "project_id", "cidr", "ip_version",
					}))
			},
			ExpectedMetrics: "",
		},
	}

	testutil.RunCollectorTests(t, tests, NewSubnetCollector)
}

func TestIPRangeSize(t *testing.T) {
	tests := []struct {
		firstIP  string
		lastIP   string
		expected int64
	}{
		{"10.13.55.230", "10.13.55.249", 20},
		{"10.0.0.1", "10.0.0.1", 1},
		{"192.168.0.0", "192.168.0.255", 256},
		{"10.0.0.2", "10.0.0.254", 253},
		{"invalid", "10.0.0.1", 0},
		{"10.0.0.1", "invalid", 0},
	}

	for _, tt := range tests {
		got := ipRangeSize(tt.firstIP, tt.lastIP)
		if got != tt.expected {
			t.Errorf("ipRangeSize(%q, %q) = %d, want %d", tt.firstIP, tt.lastIP, got, tt.expected)
		}
	}
}
