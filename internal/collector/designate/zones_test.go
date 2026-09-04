package designate

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	designatedb "github.com/vexxhost/openstack_database_exporter/internal/db/designate"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestZonesCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with zones and recordsets",
			SetupMock: func(mock sqlmock.Sqlmock) {
				zoneCols := []string{"id", "name", "tenant_id", "type", "status"}
				mock.ExpectQuery(regexp.QuoteMeta(designatedb.GetZones)).WillReturnRows(
					sqlmock.NewRows(zoneCols).
						AddRow("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "example.com.", "tenant-1", "PRIMARY", "ACTIVE").
						AddRow("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "test.com.", "tenant-2", "PRIMARY", "PENDING"),
				)

				rsCols := []string{"id", "name", "type", "zone_id", "zone_name", "tenant_id", "status"}
				mock.ExpectQuery(regexp.QuoteMeta(designatedb.GetRecordsets)).WillReturnRows(
					sqlmock.NewRows(rsCols).
						AddRow("cccccccccccccccccccccccccccccccc", "www.example.com.", "A", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "example.com.", "tenant-1", "ACTIVE").
						AddRow("dddddddddddddddddddddddddddddddd", "ns.example.com.", "NS", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "example.com.", "tenant-1", "ACTIVE").
						AddRow("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", "test.com.", "SOA", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "test.com.", "tenant-2", "PENDING"),
				)
			},
			ExpectedMetrics: `# HELP openstack_designate_recordsets recordsets
# TYPE openstack_designate_recordsets gauge
openstack_designate_recordsets{tenant_id="tenant-1",zone_id="aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",zone_name="example.com."} 2
openstack_designate_recordsets{tenant_id="tenant-2",zone_id="bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",zone_name="test.com."} 1
# HELP openstack_designate_recordsets_status recordsets_status
# TYPE openstack_designate_recordsets_status gauge
openstack_designate_recordsets_status{id="cccccccc-cccc-cccc-cccc-cccccccccccc",name="www.example.com.",status="ACTIVE",type="A",zone_id="aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",zone_name="example.com."} 1
openstack_designate_recordsets_status{id="dddddddd-dddd-dddd-dddd-dddddddddddd",name="ns.example.com.",status="ACTIVE",type="NS",zone_id="aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",zone_name="example.com."} 1
openstack_designate_recordsets_status{id="eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",name="test.com.",status="PENDING",type="SOA",zone_id="bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",zone_name="test.com."} 0
# HELP openstack_designate_up up
# TYPE openstack_designate_up gauge
openstack_designate_up 1
# HELP openstack_designate_zone_status zone_status
# TYPE openstack_designate_zone_status gauge
openstack_designate_zone_status{id="aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",name="example.com.",status="ACTIVE",tenant_id="tenant-1",type="PRIMARY"} 1
openstack_designate_zone_status{id="bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",name="test.com.",status="PENDING",tenant_id="tenant-2",type="PRIMARY"} 0
# HELP openstack_designate_zones zones
# TYPE openstack_designate_zones gauge
openstack_designate_zones 2
`,
		},
		{
			Name: "empty database",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(designatedb.GetZones)).WillReturnRows(
					sqlmock.NewRows([]string{"id", "name", "tenant_id", "type", "status"}),
				)
				mock.ExpectQuery(regexp.QuoteMeta(designatedb.GetRecordsets)).WillReturnRows(
					sqlmock.NewRows([]string{"id", "name", "type", "zone_id", "zone_name", "tenant_id", "status"}),
				)
			},
			ExpectedMetrics: `# HELP openstack_designate_up up
# TYPE openstack_designate_up gauge
openstack_designate_up 1
# HELP openstack_designate_zones zones
# TYPE openstack_designate_zones gauge
openstack_designate_zones 0
`,
		},
		{
			Name: "zone query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(designatedb.GetZones)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_designate_up up
# TYPE openstack_designate_up gauge
openstack_designate_up 0
`,
		},
	}

	testutil.RunCollectorTests(t, tests, NewZonesCollector)
}
