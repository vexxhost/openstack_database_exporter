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

func TestRegionsCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with region data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup region metrics query
				regionRows := sqlmock.NewRows([]string{
					"id", "description", "parent_region_id",
				}).AddRow(
					"RegionOne", "", "",
				)
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetRegionMetrics)).WillReturnRows(regionRows)
			},
			ExpectedMetrics: `# HELP openstack_identity_regions regions
# TYPE openstack_identity_regions gauge
openstack_identity_regions 1
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Setup empty region metrics query
				regionRows := sqlmock.NewRows([]string{
					"id", "description", "parent_region_id",
				})
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetRegionMetrics)).WillReturnRows(regionRows)
			},
			ExpectedMetrics: `# HELP openstack_identity_regions regions
# TYPE openstack_identity_regions gauge
openstack_identity_regions 0
`,
		},
		{
			Name: "database error on region query",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(keystonedb.GetRegionMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) *RegionsCollector {
		return NewRegionsCollector(db, logger)
	})
}
