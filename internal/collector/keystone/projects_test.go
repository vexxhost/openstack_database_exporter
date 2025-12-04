package keystone

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	keystonedb "github.com/vexxhost/openstack_database_exporter/internal/db/keystone"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestProjectsCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with project data",
			SetupMock: func(mock sqlmock.Sqlmock) {
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
			},
			ExpectedMetrics: `# HELP openstack_identity_project_info project_info
# TYPE openstack_identity_project_info gauge
openstack_identity_project_info{description="",domain_id="1bc2169ca88e4cdaaba46d4c15390b65",enabled="true",id="4b1eb781a47440acb8af9850103e537f",is_domain="false",name="swifttenanttest4",parent_id="",tags=""} 1
openstack_identity_project_info{description="",domain_id="default",enabled="true",id="0c4e939acacf4376bdcd1129f1a054ad",is_domain="false",name="admin",parent_id="",tags=""} 1
openstack_identity_project_info{description="Demo Project",domain_id="default",enabled="true",id="0cbd49cbf76d405d9c86562e1d579bd3",is_domain="false",name="demo",parent_id="",tags=""} 1
# HELP openstack_identity_projects projects
# TYPE openstack_identity_projects gauge
openstack_identity_projects 3
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup empty project metrics query
				projectRows := sqlmock.NewRows([]string{
					"id", "name", "description", "enabled", "domain_id", "parent_id", "is_domain", "tags",
				})
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetProjectMetrics)).WillReturnRows(projectRows)
			},
			ExpectedMetrics: `# HELP openstack_identity_projects projects
# TYPE openstack_identity_projects gauge
openstack_identity_projects 0
`,
		},
		{
			Name: "disabled project and is_domain=true",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup project metrics query with disabled project and is_domain=true
				projectRows := sqlmock.NewRows([]string{
					"id", "name", "description", "enabled", "domain_id", "parent_id", "is_domain", "tags",
				}).AddRow(
					"project-1", "project-1", "Disabled project", 0, "default", "parent-1", 1, "tag1,tag2",
				)
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetProjectMetrics)).WillReturnRows(projectRows)
			},
			ExpectedMetrics: `# HELP openstack_identity_project_info project_info
# TYPE openstack_identity_project_info gauge
openstack_identity_project_info{description="Disabled project",domain_id="default",enabled="false",id="project-1",is_domain="true",name="project-1",parent_id="parent-1",tags="tag1,tag2"} 1
# HELP openstack_identity_projects projects
# TYPE openstack_identity_projects gauge
openstack_identity_projects 1
`,
		},
		{
			Name: "null values handling",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup project metrics query with null enabled and null tags
				projectRows := sqlmock.NewRows([]string{
					"id", "name", "description", "enabled", "domain_id", "parent_id", "is_domain", "tags",
				}).AddRow(
					"project-1", "project-1", "Project description", nil, "default", "", 0, nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetProjectMetrics)).WillReturnRows(projectRows)
			},
			ExpectedMetrics: `# HELP openstack_identity_project_info project_info
# TYPE openstack_identity_project_info gauge
openstack_identity_project_info{description="Project description",domain_id="default",enabled="false",id="project-1",is_domain="false",name="project-1",parent_id="",tags=""} 1
# HELP openstack_identity_projects projects
# TYPE openstack_identity_projects gauge
openstack_identity_projects 1
`,
		},
		{
			Name: "database error on project query",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetProjectMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		return &testProjectsCollector{NewProjectsCollector(db, logger)}
	})
}

type testProjectsCollector struct {
	*ProjectsCollector
}

func (t *testProjectsCollector) Collect(ch chan<- prometheus.Metric) {
	_ = t.ProjectsCollector.Collect(ch)
}
