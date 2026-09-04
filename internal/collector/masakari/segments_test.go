package masakari

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	masakaridb "github.com/vexxhost/openstack_database_exporter/internal/db/masakari"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestSegmentsCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with segments and hosts",
			SetupMock: func(mock sqlmock.Sqlmock) {
				segmentCols := []string{"id", "uuid", "name", "service_type", "enabled", "description", "recovery_method"}
				mock.ExpectQuery(regexp.QuoteMeta(masakaridb.GetFailoverSegments)).WillReturnRows(
					sqlmock.NewRows(segmentCols).
						AddRow(1, "seg-uuid-1", "segment-a", "compute", true, "Test segment", "auto").
						AddRow(2, "seg-uuid-2", "segment-b", "compute", true, nil, "reserved_host"),
				)

				hostCols := []string{"id", "uuid", "name", "reserved", "type", "control_attributes", "on_maintenance", "failover_segment_id", "failover_segment_name"}
				mock.ExpectQuery(regexp.QuoteMeta(masakaridb.GetHosts)).WillReturnRows(
					sqlmock.NewRows(hostCols).
						AddRow(1, "host-uuid-1", "compute-1", false, "compute", "ssh", false, "seg-uuid-1", "segment-a").
						AddRow(2, "host-uuid-2", "compute-2", true, "compute", "ssh", true, "seg-uuid-1", "segment-a"),
				)

				notifCols := []string{"status", "count"}
				mock.ExpectQuery(regexp.QuoteMeta(masakaridb.GetNotificationsByStatus)).WillReturnRows(
					sqlmock.NewRows(notifCols).
						AddRow("new", 3).
						AddRow("finished", 10),
				)
			},
			ExpectedMetrics: `# HELP openstack_masakari_host host
# TYPE openstack_masakari_host gauge
openstack_masakari_host{control_attributes="ssh",failover_segment_id="seg-uuid-1",failover_segment_name="segment-a",hostname="compute-1",id="1",type="compute",uuid="host-uuid-1"} 1
openstack_masakari_host{control_attributes="ssh",failover_segment_id="seg-uuid-1",failover_segment_name="segment-a",hostname="compute-2",id="2",type="compute",uuid="host-uuid-2"} 1
# HELP openstack_masakari_host_on_maintenance host_on_maintenance
# TYPE openstack_masakari_host_on_maintenance gauge
openstack_masakari_host_on_maintenance{failover_segment_id="seg-uuid-1",hostname="compute-1",uuid="host-uuid-1"} 0
openstack_masakari_host_on_maintenance{failover_segment_id="seg-uuid-1",hostname="compute-2",uuid="host-uuid-2"} 1
# HELP openstack_masakari_host_reserved host_reserved
# TYPE openstack_masakari_host_reserved gauge
openstack_masakari_host_reserved{failover_segment_id="seg-uuid-1",hostname="compute-1",uuid="host-uuid-1"} 0
openstack_masakari_host_reserved{failover_segment_id="seg-uuid-1",hostname="compute-2",uuid="host-uuid-2"} 1
# HELP openstack_masakari_notifications notifications
# TYPE openstack_masakari_notifications gauge
openstack_masakari_notifications{status="finished"} 10
openstack_masakari_notifications{status="new"} 3
# HELP openstack_masakari_segment segment
# TYPE openstack_masakari_segment gauge
openstack_masakari_segment{description="Test segment",id="1",name="segment-a",recovery_method="auto",service_type="compute",uuid="seg-uuid-1"} 1
openstack_masakari_segment{description="",id="2",name="segment-b",recovery_method="reserved_host",service_type="compute",uuid="seg-uuid-2"} 1
# HELP openstack_masakari_up up
# TYPE openstack_masakari_up gauge
openstack_masakari_up 1
`,
		},
		{
			Name: "empty database",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(masakaridb.GetFailoverSegments)).WillReturnRows(
					sqlmock.NewRows([]string{"id", "uuid", "name", "service_type", "enabled", "description", "recovery_method"}),
				)
				mock.ExpectQuery(regexp.QuoteMeta(masakaridb.GetHosts)).WillReturnRows(
					sqlmock.NewRows([]string{"id", "uuid", "name", "reserved", "type", "control_attributes", "on_maintenance", "failover_segment_id", "failover_segment_name"}),
				)
				mock.ExpectQuery(regexp.QuoteMeta(masakaridb.GetNotificationsByStatus)).WillReturnRows(
					sqlmock.NewRows([]string{"status", "count"}),
				)
			},
			ExpectedMetrics: `# HELP openstack_masakari_up up
# TYPE openstack_masakari_up gauge
openstack_masakari_up 1
`,
		},
		{
			Name: "segment query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(masakaridb.GetFailoverSegments)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_masakari_up up
# TYPE openstack_masakari_up gauge
openstack_masakari_up 0
`,
		},
	}

	testutil.RunCollectorTests(t, tests, NewSegmentsCollector)
}
