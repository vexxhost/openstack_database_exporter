package cinder

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
	cinderdb "github.com/vexxhost/openstackdb/cinder/db"
)

func TestSnapshotsCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with snapshots",
			SetupMock: func(mock sqlmock.Sqlmock) {
				count := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.SnapshotCount)).WillReturnRows(count)
			},
			ExpectedMetrics: `# HELP openstack_cinder_snapshots snapshots
# TYPE openstack_cinder_snapshots gauge
openstack_cinder_snapshots 1
`,
		},
		{
			Name: "zero snapshots",
			SetupMock: func(mock sqlmock.Sqlmock) {
				count := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.SnapshotCount)).WillReturnRows(count)
			},
			ExpectedMetrics: `# HELP openstack_cinder_snapshots snapshots
# TYPE openstack_cinder_snapshots gauge
openstack_cinder_snapshots 0
`,
		},
		{
			Name: "large snapshot count",
			SetupMock: func(mock sqlmock.Sqlmock) {
				count := sqlmock.NewRows([]string{"count"}).AddRow(99999)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.SnapshotCount)).WillReturnRows(count)
			},
			ExpectedMetrics: `# HELP openstack_cinder_snapshots snapshots
# TYPE openstack_cinder_snapshots gauge
openstack_cinder_snapshots 99999
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.SnapshotCount)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
			ExpectError:     true,
		},
	}

	testutil.RunCollectorTests(t, tests, NewSnapshotsCollector)
}
