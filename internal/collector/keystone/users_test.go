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

func TestUsersCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with user data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup user metrics query
				userRows := sqlmock.NewRows([]string{
					"id", "enabled", "domain_id", "default_project_id", "created_at", "last_active_at",
				}).AddRow(
					"user-1", 1, "default", "0c4e939acacf4376bdcd1129f1a054ad", nil, nil,
				).AddRow(
					"user-2", 1, "default", "", nil, nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetUserMetrics)).WillReturnRows(userRows)
			},
			ExpectedMetrics: `# HELP openstack_identity_users users
# TYPE openstack_identity_users gauge
openstack_identity_users 2
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup empty user metrics query
				userRows := sqlmock.NewRows([]string{
					"id", "enabled", "domain_id", "default_project_id", "created_at", "last_active_at",
				})
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetUserMetrics)).WillReturnRows(userRows)
			},
			ExpectedMetrics: `# HELP openstack_identity_users users
# TYPE openstack_identity_users gauge
openstack_identity_users 0
`,
		},
		{
			Name: "database error on user query",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetUserMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) *UsersCollector {
		return NewUsersCollector(db, logger)
	})
}