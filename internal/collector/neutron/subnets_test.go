package neutron

import (
	"database/sql"
	"net/netip"
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

				// collectSubnetPools queries GetSubnets again
				subnetRows2 := sqlmock.NewRows([]string{
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
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnets)).WillReturnRows(subnetRows2)

				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnetPools)).WillReturnRows(
					sqlmock.NewRows([]string{
						"id", "ip_version", "max_prefixlen", "min_prefixlen",
						"default_prefixlen", "project_id", "name", "prefixes",
					}))
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
			Name: "successful collection with subnet pools",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Subnet belonging to a pool
				subnetRows := sqlmock.NewRows([]string{
					"id", "name", "cidr", "gateway_ip", "network_id",
					"project_id", "enable_dhcp", "dns_nameservers",
					"subnetpool_id", "tags",
				}).AddRow(
					"fc9e37c2-a5fd-442a-8a94-79c8351b57f0",
					"pool-subnet",
					"10.0.0.0/26",
					"10.0.0.1",
					"667a6256-d975-49ed-bfe1-822887202213",
					"d6fbbee0aa214c20b984292531ce7bd0",
					true,
					[]byte(""),
					"044ee702-b41d-4517-ac95-d0319579775b",
					[]byte(""),
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnets)).WillReturnRows(subnetRows)

				// Empty IP availability rows
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

				// collectSubnetPools: second GetSubnets call
				subnetRows2 := sqlmock.NewRows([]string{
					"id", "name", "cidr", "gateway_ip", "network_id",
					"project_id", "enable_dhcp", "dns_nameservers",
					"subnetpool_id", "tags",
				}).AddRow(
					"fc9e37c2-a5fd-442a-8a94-79c8351b57f0",
					"pool-subnet",
					"10.0.0.0/26",
					"10.0.0.1",
					"667a6256-d975-49ed-bfe1-822887202213",
					"d6fbbee0aa214c20b984292531ce7bd0",
					true,
					[]byte(""),
					"044ee702-b41d-4517-ac95-d0319579775b",
					[]byte(""),
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnets)).WillReturnRows(subnetRows2)

				// Subnet pool with prefix 10.0.0.0/24, min=24, max=26
				poolRows := sqlmock.NewRows([]string{
					"id", "ip_version", "max_prefixlen", "min_prefixlen",
					"default_prefixlen", "project_id", "name", "prefixes",
				}).AddRow(
					"044ee702-b41d-4517-ac95-d0319579775b",
					4,
					26,
					24,
					26,
					"8d652a8c66594b328c6a6bcf617aba5d",
					"shared-default-subnetpool-v4",
					"10.0.0.0/24",
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnetPools)).WillReturnRows(poolRows)
			},
			// Pool prefix 10.0.0.0/24, subnet 10.0.0.0/26 allocated.
			// For prefix lengths 24..26:
			//   total: /24=1, /25=2, /26=4
			//   used:  /24=0, /25=0, /26=1
			//   free (remaining after removing /26: {10.0.0.64/26, 10.0.0.128/25}):
			//     /24=0, /25=1, /26=3
			ExpectedMetrics: `# HELP openstack_neutron_subnet subnet
# TYPE openstack_neutron_subnet gauge
openstack_neutron_subnet{cidr="10.0.0.0/26",dns_nameservers="",enable_dhcp="true",gateway_ip="10.0.0.1",id="fc9e37c2-a5fd-442a-8a94-79c8351b57f0",name="pool-subnet",network_id="667a6256-d975-49ed-bfe1-822887202213",tags="",tenant_id="d6fbbee0aa214c20b984292531ce7bd0"} 1
# HELP openstack_neutron_subnets subnets
# TYPE openstack_neutron_subnets gauge
openstack_neutron_subnets 1
# HELP openstack_neutron_subnets_free subnets_free
# TYPE openstack_neutron_subnets_free gauge
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/24",prefix_length="24",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/24",prefix_length="25",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 1
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/24",prefix_length="26",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 3
# HELP openstack_neutron_subnets_total subnets_total
# TYPE openstack_neutron_subnets_total gauge
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/24",prefix_length="24",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 1
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/24",prefix_length="25",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 2
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/24",prefix_length="26",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 4
# HELP openstack_neutron_subnets_used subnets_used
# TYPE openstack_neutron_subnets_used gauge
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/24",prefix_length="24",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/24",prefix_length="25",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/24",prefix_length="26",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 1
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
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnets)).WillReturnRows(
					sqlmock.NewRows([]string{
						"id", "name", "cidr", "gateway_ip", "network_id",
						"project_id", "enable_dhcp", "dns_nameservers",
						"subnetpool_id", "tags",
					}))
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnetPools)).WillReturnRows(
					sqlmock.NewRows([]string{
						"id", "ip_version", "max_prefixlen", "min_prefixlen",
						"default_prefixlen", "project_id", "name", "prefixes",
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
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnets)).WillReturnError(sql.ErrConnDone)
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

func TestCalculateFreeSubnets(t *testing.T) {
	tests := []struct {
		name         string
		poolPrefix   string
		subnets      []string
		prefixLength int
		expected     float64
	}{
		{
			name:         "no subnets allocated in /24 pool at /26",
			poolPrefix:   "10.0.0.0/24",
			subnets:      nil,
			prefixLength: 26,
			expected:     4,
		},
		{
			name:         "one /26 allocated from /24 pool at /26",
			poolPrefix:   "10.0.0.0/24",
			subnets:      []string{"10.0.0.0/26"},
			prefixLength: 26,
			expected:     3,
		},
		{
			name:         "one /26 allocated from /24 pool at /25",
			poolPrefix:   "10.0.0.0/24",
			subnets:      []string{"10.0.0.0/26"},
			prefixLength: 25,
			expected:     1,
		},
		{
			name:         "one /26 allocated from /24 pool at /24",
			poolPrefix:   "10.0.0.0/24",
			subnets:      []string{"10.0.0.0/26"},
			prefixLength: 24,
			expected:     0,
		},
		{
			name:         "fully allocated pool",
			poolPrefix:   "10.0.0.0/24",
			subnets:      []string{"10.0.0.0/24"},
			prefixLength: 24,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := netip.MustParsePrefix(tt.poolPrefix)
			var subnets []netip.Prefix
			for _, s := range tt.subnets {
				subnets = append(subnets, netip.MustParsePrefix(s))
			}
			got, err := calculateFreeSubnets(&prefix, subnets, tt.prefixLength)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("calculateFreeSubnets(%q, %v, %d) = %v, want %v",
					tt.poolPrefix, tt.subnets, tt.prefixLength, got, tt.expected)
			}
		})
	}
}

func TestCalculateUsedSubnets(t *testing.T) {
	tests := []struct {
		name         string
		subnets      []string
		ipPrefix     string
		prefixLength int
		expected     float64
	}{
		{
			name:         "no subnets",
			subnets:      nil,
			ipPrefix:     "10.0.0.0/24",
			prefixLength: 26,
			expected:     0,
		},
		{
			name:         "one /26 in pool counted at /26",
			subnets:      []string{"10.0.0.0/26"},
			ipPrefix:     "10.0.0.0/24",
			prefixLength: 26,
			expected:     1,
		},
		{
			name:         "one /26 in pool not counted at /25",
			subnets:      []string{"10.0.0.0/26"},
			ipPrefix:     "10.0.0.0/24",
			prefixLength: 25,
			expected:     0,
		},
		{
			name:         "two /26 subnets counted at /26",
			subnets:      []string{"10.0.0.0/26", "10.0.0.64/26"},
			ipPrefix:     "10.0.0.0/24",
			prefixLength: 26,
			expected:     2,
		},
		{
			name:         "non-overlapping subnet ignored",
			subnets:      []string{"192.168.0.0/24"},
			ipPrefix:     "10.0.0.0/24",
			prefixLength: 24,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var subnets []netip.Prefix
			for _, s := range tt.subnets {
				subnets = append(subnets, netip.MustParsePrefix(s))
			}
			prefix := netip.MustParsePrefix(tt.ipPrefix)
			got := calculateUsedSubnets(subnets, prefix, tt.prefixLength)
			if got != tt.expected {
				t.Errorf("calculateUsedSubnets(%v, %q, %d) = %v, want %v",
					tt.subnets, tt.ipPrefix, tt.prefixLength, got, tt.expected)
			}
		})
	}
}
