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
				}).AddRow(
					"fc9e37c2-a5fd-442a-8a94-79c8351b57f0",
					"10.0.0.0/26",
					"10.0.0.1",
					"667a6256-d975-49ed-bfe1-822887202213",
					"d6fbbee0aa214c20b984292531ce7bd0",
					"true",
					"",
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnets)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_subnet subnet
# TYPE openstack_neutron_subnet gauge
openstack_neutron_subnet{cidr="10.0.0.0/26",dns_nameservers="",enable_dhcp="true",gateway_ip="10.0.0.1",id="fc9e37c2-a5fd-442a-8a94-79c8351b57f0",network_id="667a6256-d975-49ed-bfe1-822887202213",project_id="d6fbbee0aa214c20b984292531ce7bd0"} 1
# HELP openstack_neutron_subnets subnets
# TYPE openstack_neutron_subnets gauge
openstack_neutron_subnets 1
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
