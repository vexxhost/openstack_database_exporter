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

func TestNetworkCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection of networks",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id",
					"name",
					"project_id",
					"status",
					"provider_network_type",
					"provider_physical_network",
					"provider_segmentation_id",
					"subnets",
					"is_external",
					"is_shared",
				}).AddRow(
					"8b0c67b2-4ca7-40d4-9025-8c07c5cbc47b",
					"shared",
					"642b6bb02dc94292ad235a607735a791",
					"ACTIVE",
					"geneve",
					"",
					"22054",
					"eda4ffcb-2afd-40e9-be82-0e7783093a3f",
					0,
					1,
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetNetworks)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_network network
# TYPE openstack_neutron_network gauge
openstack_neutron_network{id="8b0c67b2-4ca7-40d4-9025-8c07c5cbc47b",is_external="false",is_shared="true",name="shared",project_id="642b6bb02dc94292ad235a607735a791",provider_network_type="geneve",provider_physical_network="",provider_segmentation_id="22054",status="ACTIVE",subnets="eda4ffcb-2afd-40e9-be82-0e7783093a3f"} 1
# HELP openstack_neutron_networks networks
# TYPE openstack_neutron_networks gauge
openstack_neutron_networks 1
`,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		return &testNetworkCollector{NewNetworkCollector(db, logger)}
	})
}

type testNetworkCollector struct {
	*NetworkCollector
}

func (t *testNetworkCollector) Collect(ch chan<- prometheus.Metric) {
	if err := t.NetworkCollector.Collect(ch); err != nil {
		panic("unexpected error: " + err.Error())
	}
}
