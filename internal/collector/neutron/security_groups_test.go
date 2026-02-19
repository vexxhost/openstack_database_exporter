package neutron

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestSecurityGroupCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful count",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"cnt"}).AddRow(16)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSecurityGroupCount)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_security_groups security_groups
# TYPE openstack_neutron_security_groups gauge
openstack_neutron_security_groups 16
`,
		},
		{
			Name: "zero count",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"cnt"}).AddRow(0)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSecurityGroupCount)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_security_groups security_groups
# TYPE openstack_neutron_security_groups gauge
openstack_neutron_security_groups 0
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetSecurityGroupCount)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
		},
	}

	testutil.RunCollectorTests(t, tests, NewSecurityGroupCollector)
}
