package manila

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	maniladb "github.com/vexxhost/openstack_database_exporter/internal/db/manila"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestSharesCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with manila shares",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "project_id", "size", "share_proto", "status", "share_type_name", "availability_zone",
				}).AddRow(
					"4be93e2e-ffff-ffff-ffff-603e3ec2a5d6", "share-test", "ffff8fa0ca1a468db8ad00970c1effff", 1, "NFS", "available", "", "az1",
				)

				mock.ExpectQuery(regexp.QuoteMeta(maniladb.GetShareMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_sharev2_share_gb share_gb
# TYPE openstack_sharev2_share_gb gauge
openstack_sharev2_share_gb{availability_zone="az1",id="4be93e2e-ffff-ffff-ffff-603e3ec2a5d6",name="share-test",project_id="ffff8fa0ca1a468db8ad00970c1effff",share_proto="NFS",share_type="az1",share_type_name="",status="available"} 1
# HELP openstack_sharev2_share_status share_status
# TYPE openstack_sharev2_share_status gauge
openstack_sharev2_share_status{id="4be93e2e-ffff-ffff-ffff-603e3ec2a5d6",name="share-test",project_id="ffff8fa0ca1a468db8ad00970c1effff",share_proto="NFS",share_type="az1",share_type_name="",size="1",status="available"} 1
# HELP openstack_sharev2_share_status_counter share_status_counter
# TYPE openstack_sharev2_share_status_counter gauge
openstack_sharev2_share_status_counter{status="available"} 1
openstack_sharev2_share_status_counter{status="creating"} 0
openstack_sharev2_share_status_counter{status="deleting"} 0
openstack_sharev2_share_status_counter{status="error"} 0
openstack_sharev2_share_status_counter{status="error_deleting"} 0
openstack_sharev2_share_status_counter{status="extending"} 0
openstack_sharev2_share_status_counter{status="inactive"} 0
openstack_sharev2_share_status_counter{status="managing"} 0
openstack_sharev2_share_status_counter{status="migrating"} 0
openstack_sharev2_share_status_counter{status="migration_error"} 0
openstack_sharev2_share_status_counter{status="restoring"} 0
openstack_sharev2_share_status_counter{status="reverting"} 0
openstack_sharev2_share_status_counter{status="reverting_error"} 0
openstack_sharev2_share_status_counter{status="reverting_to_snapshot"} 0
openstack_sharev2_share_status_counter{status="shrinking"} 0
openstack_sharev2_share_status_counter{status="shrinking_error"} 0
openstack_sharev2_share_status_counter{status="soft_deleting"} 0
openstack_sharev2_share_status_counter{status="unmanaging"} 0
openstack_sharev2_share_status_counter{status="updating"} 0
# HELP openstack_sharev2_shares_counter shares_counter
# TYPE openstack_sharev2_shares_counter gauge
openstack_sharev2_shares_counter 1
# HELP openstack_sharev2_up up
# TYPE openstack_sharev2_up gauge
openstack_sharev2_up 1
`,
		},
		{
			Name: "successful collection with multiple shares",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "project_id", "size", "share_proto", "status", "share_type_name", "availability_zone",
				}).AddRow(
					"share-1", "test-share-1", "project-1", 10, "NFS", "available", "default", "nova",
				).AddRow(
					"share-2", "test-share-2", "project-2", 20, "CIFS", "creating", "ssd", "nova",
				).AddRow(
					"share-3", "test-share-3", "project-1", 5, "NFS", "error", "default", "nova",
				)

				mock.ExpectQuery(regexp.QuoteMeta(maniladb.GetShareMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_sharev2_share_gb share_gb
# TYPE openstack_sharev2_share_gb gauge
openstack_sharev2_share_gb{availability_zone="nova",id="share-1",name="test-share-1",project_id="project-1",share_proto="NFS",share_type="nova",share_type_name="default",status="available"} 10
openstack_sharev2_share_gb{availability_zone="nova",id="share-2",name="test-share-2",project_id="project-2",share_proto="CIFS",share_type="nova",share_type_name="ssd",status="creating"} 20
openstack_sharev2_share_gb{availability_zone="nova",id="share-3",name="test-share-3",project_id="project-1",share_proto="NFS",share_type="nova",share_type_name="default",status="error"} 5
# HELP openstack_sharev2_share_status share_status
# TYPE openstack_sharev2_share_status gauge
openstack_sharev2_share_status{id="share-1",name="test-share-1",project_id="project-1",share_proto="NFS",share_type="nova",share_type_name="default",size="10",status="available"} 1
openstack_sharev2_share_status{id="share-2",name="test-share-2",project_id="project-2",share_proto="CIFS",share_type="nova",share_type_name="ssd",size="20",status="creating"} 1
openstack_sharev2_share_status{id="share-3",name="test-share-3",project_id="project-1",share_proto="NFS",share_type="nova",share_type_name="default",size="5",status="error"} 1
# HELP openstack_sharev2_share_status_counter share_status_counter
# TYPE openstack_sharev2_share_status_counter gauge
openstack_sharev2_share_status_counter{status="available"} 1
openstack_sharev2_share_status_counter{status="creating"} 1
openstack_sharev2_share_status_counter{status="deleting"} 0
openstack_sharev2_share_status_counter{status="error"} 1
openstack_sharev2_share_status_counter{status="error_deleting"} 0
openstack_sharev2_share_status_counter{status="extending"} 0
openstack_sharev2_share_status_counter{status="inactive"} 0
openstack_sharev2_share_status_counter{status="managing"} 0
openstack_sharev2_share_status_counter{status="migrating"} 0
openstack_sharev2_share_status_counter{status="migration_error"} 0
openstack_sharev2_share_status_counter{status="restoring"} 0
openstack_sharev2_share_status_counter{status="reverting"} 0
openstack_sharev2_share_status_counter{status="reverting_error"} 0
openstack_sharev2_share_status_counter{status="reverting_to_snapshot"} 0
openstack_sharev2_share_status_counter{status="shrinking"} 0
openstack_sharev2_share_status_counter{status="shrinking_error"} 0
openstack_sharev2_share_status_counter{status="soft_deleting"} 0
openstack_sharev2_share_status_counter{status="unmanaging"} 0
openstack_sharev2_share_status_counter{status="updating"} 0
# HELP openstack_sharev2_shares_counter shares_counter
# TYPE openstack_sharev2_shares_counter gauge
openstack_sharev2_shares_counter 3
# HELP openstack_sharev2_up up
# TYPE openstack_sharev2_up gauge
openstack_sharev2_up 1
`,
		},
		{
			Name: "successful collection with no shares",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "project_id", "size", "share_proto", "status", "share_type_name", "availability_zone",
				})

				mock.ExpectQuery(regexp.QuoteMeta(maniladb.GetShareMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_sharev2_share_status_counter share_status_counter
# TYPE openstack_sharev2_share_status_counter gauge
openstack_sharev2_share_status_counter{status="available"} 0
openstack_sharev2_share_status_counter{status="creating"} 0
openstack_sharev2_share_status_counter{status="deleting"} 0
openstack_sharev2_share_status_counter{status="error"} 0
openstack_sharev2_share_status_counter{status="error_deleting"} 0
openstack_sharev2_share_status_counter{status="extending"} 0
openstack_sharev2_share_status_counter{status="inactive"} 0
openstack_sharev2_share_status_counter{status="managing"} 0
openstack_sharev2_share_status_counter{status="migrating"} 0
openstack_sharev2_share_status_counter{status="migration_error"} 0
openstack_sharev2_share_status_counter{status="restoring"} 0
openstack_sharev2_share_status_counter{status="reverting"} 0
openstack_sharev2_share_status_counter{status="reverting_error"} 0
openstack_sharev2_share_status_counter{status="reverting_to_snapshot"} 0
openstack_sharev2_share_status_counter{status="shrinking"} 0
openstack_sharev2_share_status_counter{status="shrinking_error"} 0
openstack_sharev2_share_status_counter{status="soft_deleting"} 0
openstack_sharev2_share_status_counter{status="unmanaging"} 0
openstack_sharev2_share_status_counter{status="updating"} 0
# HELP openstack_sharev2_shares_counter shares_counter
# TYPE openstack_sharev2_shares_counter gauge
openstack_sharev2_shares_counter 0
# HELP openstack_sharev2_up up
# TYPE openstack_sharev2_up gauge
openstack_sharev2_up 1
`,
		},
		{
			Name: "handles null values gracefully",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "project_id", "size", "share_proto", "status", "share_type_name", "availability_zone",
				}).AddRow(
					"share-null", sql.NullString{Valid: false}, sql.NullString{Valid: false}, sql.NullInt32{Valid: false}, sql.NullString{Valid: false}, sql.NullString{Valid: false}, "", "",
				)

				mock.ExpectQuery(regexp.QuoteMeta(maniladb.GetShareMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_sharev2_share_gb share_gb
# TYPE openstack_sharev2_share_gb gauge
openstack_sharev2_share_gb{availability_zone="",id="share-null",name="",project_id="",share_proto="",share_type="",share_type_name="",status=""} 0
# HELP openstack_sharev2_share_status share_status
# TYPE openstack_sharev2_share_status gauge
openstack_sharev2_share_status{id="share-null",name="",project_id="",share_proto="",share_type="",share_type_name="",size="0",status=""} 0
# HELP openstack_sharev2_share_status_counter share_status_counter
# TYPE openstack_sharev2_share_status_counter gauge
openstack_sharev2_share_status_counter{status="available"} 0
openstack_sharev2_share_status_counter{status="creating"} 0
openstack_sharev2_share_status_counter{status="deleting"} 0
openstack_sharev2_share_status_counter{status="error"} 0
openstack_sharev2_share_status_counter{status="error_deleting"} 0
openstack_sharev2_share_status_counter{status="extending"} 0
openstack_sharev2_share_status_counter{status="inactive"} 0
openstack_sharev2_share_status_counter{status="managing"} 0
openstack_sharev2_share_status_counter{status="migrating"} 0
openstack_sharev2_share_status_counter{status="migration_error"} 0
openstack_sharev2_share_status_counter{status="restoring"} 0
openstack_sharev2_share_status_counter{status="reverting"} 0
openstack_sharev2_share_status_counter{status="reverting_error"} 0
openstack_sharev2_share_status_counter{status="reverting_to_snapshot"} 0
openstack_sharev2_share_status_counter{status="shrinking"} 0
openstack_sharev2_share_status_counter{status="shrinking_error"} 0
openstack_sharev2_share_status_counter{status="soft_deleting"} 0
openstack_sharev2_share_status_counter{status="unmanaging"} 0
openstack_sharev2_share_status_counter{status="updating"} 0
# HELP openstack_sharev2_shares_counter shares_counter
# TYPE openstack_sharev2_shares_counter gauge
openstack_sharev2_shares_counter 1
# HELP openstack_sharev2_up up
# TYPE openstack_sharev2_up gauge
openstack_sharev2_up 1
`,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(maniladb.GetShareMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_sharev2_up up
# TYPE openstack_sharev2_up gauge
openstack_sharev2_up 0
`,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) *SharesCollector {
		return NewSharesCollector(db, logger)
	})
}