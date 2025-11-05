package keystone

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	keystonedb "github.com/vexxhost/openstack_database_exporter/internal/db/keystone"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestGroupsCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with group data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup group metrics query
				groupRows := sqlmock.NewRows([]string{
					"id", "domain_id", "name", "description",
				}).AddRow(
					"group-1", "default", "test-group-1", "Test group 1",
				).AddRow(
					"group-2", "default", "test-group-2", "Test group 2",
				)
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetGroupMetrics)).WillReturnRows(groupRows)
			},
			ExpectedMetrics: `# HELP openstack_identity_groups groups
# TYPE openstack_identity_groups gauge
openstack_identity_groups 2
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup empty group metrics query
				groupRows := sqlmock.NewRows([]string{
					"id", "domain_id", "name", "description",
				})
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetGroupMetrics)).WillReturnRows(groupRows)
			},
			ExpectedMetrics: `# HELP openstack_identity_groups groups
# TYPE openstack_identity_groups gauge
openstack_identity_groups 0
`,
		},
		{
			Name: "database error on group query",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetGroupMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) *GroupsCollector {
		return NewGroupsCollector(db, logger)
	})
}
