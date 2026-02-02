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

func TestSecurityGroupCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection of security groups",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id",
				}).AddRow(
					"some_id_1",
				).AddRow(
					"some_id_2",
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSecurityGroups)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_security_groups security_groups
# TYPE openstack_neutron_security_groups gauge
openstack_neutron_security_groups 2
`,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		return &testSecurityGroupCollector{NewSecurityGroupCollector(db, logger)}
	})
}

type testSecurityGroupCollector struct {
	*SecurityGroupCollector
}

func (t *testSecurityGroupCollector) Collect(ch chan<- prometheus.Metric) {
	if err := t.SecurityGroupCollector.Collect(ch); err != nil {
		panic("unexpected error: " + err.Error())
	}
}
