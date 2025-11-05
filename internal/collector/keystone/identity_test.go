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

func TestIdentityCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with all keystone data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup domain metrics query
				domainRows := sqlmock.NewRows([]string{
					"id", "name", "description", "enabled",
				}).AddRow(
					"default", "Default", "Owns users and tenants (i.e. projects) available on Identity API v2.", 1,
				)
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetDomainMetrics)).WillReturnRows(domainRows)

				// Setup project metrics query
				projectRows := sqlmock.NewRows([]string{
					"id", "name", "description", "enabled", "domain_id", "parent_id", "is_domain", "tags",
				}).AddRow(
					"4b1eb781a47440acb8af9850103e537f", "swifttenanttest4", "", 1, "1bc2169ca88e4cdaaba46d4c15390b65", "", 0, "",
				).AddRow(
					"0c4e939acacf4376bdcd1129f1a054ad", "admin", "", 1, "default", "", 0, "",
				).AddRow(
					"0cbd49cbf76d405d9c86562e1d579bd3", "demo", "Demo Project", 1, "default", "", 0, "",
				)
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetProjectMetrics)).WillReturnRows(projectRows)

				// Setup group metrics query
				groupRows := sqlmock.NewRows([]string{
					"id", "domain_id", "name", "description",
				}).AddRow(
					"group-1", "default", "test-group-1", "Test group 1",
				).AddRow(
					"group-2", "default", "test-group-2", "Test group 2",
				)
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetGroupMetrics)).WillReturnRows(groupRows)

				// Setup region metrics query
				regionRows := sqlmock.NewRows([]string{
					"id", "description", "parent_region_id",
				}).AddRow(
					"RegionOne", "", "",
				)
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetRegionMetrics)).WillReturnRows(regionRows)

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
			ExpectedMetrics: `# HELP openstack_identity_domain_info domain_info
# TYPE openstack_identity_domain_info gauge
openstack_identity_domain_info{description="Owns users and tenants (i.e. projects) available on Identity API v2.",enabled="true",id="default",name="Default"} 1
# HELP openstack_identity_domains domains
# TYPE openstack_identity_domains gauge
openstack_identity_domains 1
# HELP openstack_identity_groups groups
# TYPE openstack_identity_groups gauge
openstack_identity_groups 2
# HELP openstack_identity_project_info project_info
# TYPE openstack_identity_project_info gauge
openstack_identity_project_info{description="",domain_id="1bc2169ca88e4cdaaba46d4c15390b65",enabled="true",id="4b1eb781a47440acb8af9850103e537f",is_domain="false",name="swifttenanttest4",parent_id="",tags=""} 1
openstack_identity_project_info{description="",domain_id="default",enabled="true",id="0c4e939acacf4376bdcd1129f1a054ad",is_domain="false",name="admin",parent_id="",tags=""} 1
openstack_identity_project_info{description="Demo Project",domain_id="default",enabled="true",id="0cbd49cbf76d405d9c86562e1d579bd3",is_domain="false",name="demo",parent_id="",tags=""} 1
# HELP openstack_identity_projects projects
# TYPE openstack_identity_projects gauge
openstack_identity_projects 3
# HELP openstack_identity_regions regions
# TYPE openstack_identity_regions gauge
openstack_identity_regions 1
# HELP openstack_identity_up up
# TYPE openstack_identity_up gauge
openstack_identity_up 1
# HELP openstack_identity_users users
# TYPE openstack_identity_users gauge
openstack_identity_users 2
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup empty domain metrics query
				domainRows := sqlmock.NewRows([]string{
					"id", "name", "description", "enabled",
				})
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetDomainMetrics)).WillReturnRows(domainRows)

				// Setup empty project metrics query
				projectRows := sqlmock.NewRows([]string{
					"id", "name", "description", "enabled", "domain_id", "parent_id", "is_domain", "tags",
				})
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetProjectMetrics)).WillReturnRows(projectRows)

				// Setup empty group metrics query
				groupRows := sqlmock.NewRows([]string{
					"id", "domain_id", "name", "description",
				})
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetGroupMetrics)).WillReturnRows(groupRows)

				// Setup empty region metrics query
				regionRows := sqlmock.NewRows([]string{
					"id", "description", "parent_region_id",
				})
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetRegionMetrics)).WillReturnRows(regionRows)

				// Setup empty user metrics query
				userRows := sqlmock.NewRows([]string{
					"id", "enabled", "domain_id", "default_project_id", "created_at", "last_active_at",
				})
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetUserMetrics)).WillReturnRows(userRows)
			},
			ExpectedMetrics: `# HELP openstack_identity_domains domains
# TYPE openstack_identity_domains gauge
openstack_identity_domains 0
# HELP openstack_identity_groups groups
# TYPE openstack_identity_groups gauge
openstack_identity_groups 0
# HELP openstack_identity_projects projects
# TYPE openstack_identity_projects gauge
openstack_identity_projects 0
# HELP openstack_identity_regions regions
# TYPE openstack_identity_regions gauge
openstack_identity_regions 0
# HELP openstack_identity_up up
# TYPE openstack_identity_up gauge
openstack_identity_up 1
# HELP openstack_identity_users users
# TYPE openstack_identity_users gauge
openstack_identity_users 0
`,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) *IdentityCollector {
		return NewIdentityCollector(db, logger)
	})
}
