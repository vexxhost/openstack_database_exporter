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

func TestDomainsCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with domain data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup domain metrics query
				domainRows := sqlmock.NewRows([]string{
					"id", "name", "description", "enabled",
				}).AddRow(
					"default", "Default", "Owns users and tenants (i.e. projects) available on Identity API v2.", 1,
				)
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetDomainMetrics)).WillReturnRows(domainRows)
			},
			ExpectedMetrics: `# HELP openstack_identity_domain_info domain_info
# TYPE openstack_identity_domain_info gauge
openstack_identity_domain_info{description="Owns users and tenants (i.e. projects) available on Identity API v2.",enabled="true",id="default",name="Default"} 1
# HELP openstack_identity_domains domains
# TYPE openstack_identity_domains gauge
openstack_identity_domains 1
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
			},
			ExpectedMetrics: `# HELP openstack_identity_domains domains
# TYPE openstack_identity_domains gauge
openstack_identity_domains 0
`,
		},
		{
			Name: "disabled domain",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup domain metrics query with disabled domain
				domainRows := sqlmock.NewRows([]string{
					"id", "name", "description", "enabled",
				}).AddRow(
					"disabled-domain", "Disabled Domain", "A disabled domain", 0,
				)
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetDomainMetrics)).WillReturnRows(domainRows)
			},
			ExpectedMetrics: `# HELP openstack_identity_domain_info domain_info
# TYPE openstack_identity_domain_info gauge
openstack_identity_domain_info{description="A disabled domain",enabled="false",id="disabled-domain",name="Disabled Domain"} 1
# HELP openstack_identity_domains domains
# TYPE openstack_identity_domains gauge
openstack_identity_domains 1
`,
		},
		{
			Name: "null values handling",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup domain metrics query with null enabled
				domainRows := sqlmock.NewRows([]string{
					"id", "name", "description", "enabled",
				}).AddRow(
					"domain-1", "Domain 1", "Domain description", nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetDomainMetrics)).WillReturnRows(domainRows)
			},
			ExpectedMetrics: `# HELP openstack_identity_domain_info domain_info
# TYPE openstack_identity_domain_info gauge
openstack_identity_domain_info{description="Domain description",enabled="false",id="domain-1",name="Domain 1"} 1
# HELP openstack_identity_domains domains
# TYPE openstack_identity_domains gauge
openstack_identity_domains 1
`,
		},
		{
			Name: "database error on domain query",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetDomainMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) *DomainsCollector {
		return NewDomainsCollector(db, logger)
	})
}
