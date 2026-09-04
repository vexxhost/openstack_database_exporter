package designate

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestHealthCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(healthQuery)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			ExpectedMetrics: `# HELP openstack_designate_up up
# TYPE openstack_designate_up gauge
openstack_designate_up 1
`,
		},
		{
			Name: "query failure",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(healthQuery)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_designate_up up
# TYPE openstack_designate_up gauge
openstack_designate_up 0
`,
		},
	}

	testutil.RunCollectorTests(t, tests, NewHealthCollector)
}
