package neutron

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestNetworkingCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "all collectors fail, only up metric with value 0",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetNetworks)).WillReturnError(sql.ErrConnDone)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetFloatingIPs)).WillReturnError(sql.ErrConnDone)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetRouters)).WillReturnError(sql.ErrConnDone)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetPorts)).WillReturnError(sql.ErrConnDone)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSecurityGroups)).WillReturnError(sql.ErrConnDone)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSubnets)).WillReturnError(sql.ErrConnDone)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetHARouterAgentPortBindingsWithAgents)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_neutron_up up
# TYPE openstack_neutron_up gauge
openstack_neutron_up 0
`,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) *NetworkingCollector {
		return NewNetworkingCollector(db, logger)
	})
}
