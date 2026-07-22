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

func TestProjectTagsCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with project tags",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"project_id", "name",
				}).AddRow(
					"proj-001", "env:prod",
				).AddRow(
					"proj-001", "team:infra",
				).AddRow(
					"proj-002", "env:dev",
				)
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetProjectTags)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_identity_project_tag project_tag
# TYPE openstack_identity_project_tag gauge
openstack_identity_project_tag{project_id="proj-001",tag="env:prod"} 1
openstack_identity_project_tag{project_id="proj-001",tag="team:infra"} 1
openstack_identity_project_tag{project_id="proj-002",tag="env:dev"} 1
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"project_id", "name",
				})
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetProjectTags)).WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "database error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetProjectTags)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		return &testProjectTagsCollector{NewProjectTagsCollector(db, logger)}
	})
}

type testProjectTagsCollector struct {
	*ProjectTagsCollector
}

func (t *testProjectTagsCollector) Collect(ch chan<- prometheus.Metric) {
	_ = t.ProjectTagsCollector.Collect(ch)
}
