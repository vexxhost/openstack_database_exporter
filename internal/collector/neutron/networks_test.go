package neutron

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestNetworkCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with networks",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "project_id", "status",
					"provider_network_type", "provider_physical_network",
					"provider_segmentation_id", "subnets",
					"is_external", "is_shared", "tags",
				}).AddRow(
					"6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",
					"public",
					"da457edfad314ed98fc84ef5e7d37f37",
					"ACTIVE",
					"flat",
					"external",
					[]byte(""),
					[]byte("5b32ccf9-ddbe-402b-9b68-bc66cf3c20ce"),
					1, 0,
					[]byte(""),
				)

				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetNetworks)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_network network
# TYPE openstack_neutron_network gauge
openstack_neutron_network{id="6c0ae7af-cdef-4450-b607-0c3f4c9bb10a",is_external="true",is_shared="false",name="public",provider_network_type="flat",provider_physical_network="external",provider_segmentation_id="",status="ACTIVE",subnets="5b32ccf9-ddbe-402b-9b68-bc66cf3c20ce",tags="",tenant_id="da457edfad314ed98fc84ef5e7d37f37"} 0
# HELP openstack_neutron_networks networks
# TYPE openstack_neutron_networks gauge
openstack_neutron_networks 1
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "project_id", "status",
					"provider_network_type", "provider_physical_network",
					"provider_segmentation_id", "subnets",
					"is_external", "is_shared", "tags",
				})
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetNetworks)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_networks networks
# TYPE openstack_neutron_networks gauge
openstack_neutron_networks 0
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetNetworks)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
		},
	}

	testutil.RunCollectorTests(t, tests, NewNetworkCollector)
}
