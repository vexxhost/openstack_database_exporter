package neutron

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/DATA-DOG/go-sqlmock"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestSubnetCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection of subnets",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id",
					"cidr",
					"gateway_ip",
					"network_id",
					"project_id",
					"enable_dhcp",
					"dns_nameservers",
					"subnetpool_id",
				}).AddRow(
					"fc9e37c2-a5fd-442a-8a94-79c8351b57f0",
					"10.0.0.0/26",
					"10.0.0.1",
					"667a6256-d975-49ed-bfe1-822887202213",
					"d6fbbee0aa214c20b984292531ce7bd0",
					"true",
					"",
					"",
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnets)).WillReturnRows(rows)

				poolRows := sqlmock.NewRows([]string{
					"id",
					"ip_version",
					"max_prefixlen",
					"min_prefixlen",
					"default_prefixlen",
					"project_id",
					"name",
					"prefixes",
				}).AddRow(
					"044ee702-b41d-4517-ac95-d0319579775b",
					4,
					32,
					8,
					26,
					"8d652a8c66594b328c6a6bcf617aba5d",
					"shared-default-subnetpool-v4",
					"10.0.0.0/22",
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnetPools)).WillReturnRows(poolRows)

			},
			ExpectedMetrics: `# HELP openstack_neutron_subnet subnet
# TYPE openstack_neutron_subnet gauge
openstack_neutron_subnet{cidr="10.0.0.0/26",dns_nameservers="",enable_dhcp="true",gateway_ip="10.0.0.1",id="fc9e37c2-a5fd-442a-8a94-79c8351b57f0",network_id="667a6256-d975-49ed-bfe1-822887202213",project_id="d6fbbee0aa214c20b984292531ce7bd0"} 1
# HELP openstack_neutron_subnets subnets
# TYPE openstack_neutron_subnets gauge
openstack_neutron_subnets 1
# HELP openstack_neutron_subnets_free subnets_free
# TYPE openstack_neutron_subnets_free gauge
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/22",prefix_length="22",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 1
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/22",prefix_length="23",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 2
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/22",prefix_length="24",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 4
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/22",prefix_length="25",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 8
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/22",prefix_length="26",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 16
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/22",prefix_length="27",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 32
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/22",prefix_length="28",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 64
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/22",prefix_length="29",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 128
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/22",prefix_length="30",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 256
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/22",prefix_length="31",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 512
openstack_neutron_subnets_free{ip_version="4",prefix="10.0.0.0/22",prefix_length="32",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 1024
# HELP openstack_neutron_subnets_total subnets_total
# TYPE openstack_neutron_subnets_total gauge
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/22",prefix_length="22",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 1
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/22",prefix_length="23",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 2
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/22",prefix_length="24",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 4
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/22",prefix_length="25",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 8
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/22",prefix_length="26",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 16
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/22",prefix_length="27",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 32
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/22",prefix_length="28",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 64
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/22",prefix_length="29",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 128
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/22",prefix_length="30",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 256
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/22",prefix_length="31",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 512
openstack_neutron_subnets_total{ip_version="4",prefix="10.0.0.0/22",prefix_length="32",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 1024
# HELP openstack_neutron_subnets_used subnets_used
# TYPE openstack_neutron_subnets_used gauge
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/22",prefix_length="22",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/22",prefix_length="23",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/22",prefix_length="24",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/22",prefix_length="25",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/22",prefix_length="26",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/22",prefix_length="27",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/22",prefix_length="28",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/22",prefix_length="29",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/22",prefix_length="30",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/22",prefix_length="31",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
openstack_neutron_subnets_used{ip_version="4",prefix="10.0.0.0/22",prefix_length="32",project_id="8d652a8c66594b328c6a6bcf617aba5d",subnet_pool_id="044ee702-b41d-4517-ac95-d0319579775b",subnet_pool_name="shared-default-subnetpool-v4"} 0
`,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		return &testSubnetCollector{NewSubnetCollector(db, logger)}
	})
}

type testSubnetCollector struct {
	*SubnetCollector
}

func (t *testSubnetCollector) Collect(ch chan<- prometheus.Metric) {
	if err := t.SubnetCollector.Collect(ch); err != nil {
		panic("unexpected error: " + err.Error())
	}
}
