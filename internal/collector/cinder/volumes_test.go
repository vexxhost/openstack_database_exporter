package cinder

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	cinderdb "github.com/vexxhost/openstack_database_exporter/internal/db/cinder"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestVolumesCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with volumes",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "size", "status", "availability_zone",
					"bootable", "project_id", "user_id", "volume_type", "server_id",
				}).AddRow(
					"173f7b48-c4c1-4e70-9acc-086b39073506", "test-volume", 1, "available", "nova",
					true, "bab7d5c60cd041a0a36f7c4b6e1dd978", "32779452fcd34ae1a53a797ac8a1e064", "lvmdriver-1", nil,
				).AddRow(
					"6edbc2f4-1507-44f8-ac0d-eed1d2608d38", "test-volume-attachments", 2, "in-use", "nova",
					false, "bab7d5c60cd041a0a36f7c4b6e1dd978", "32779452fcd34ae1a53a797ac8a1e064", "lvmdriver-1", "f4fda93b-06e0-4743-8117-bc8bcecd651b",
				)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetAllVolumes)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_cinder_up up
# TYPE openstack_cinder_up gauge
openstack_cinder_up 1
# HELP openstack_cinder_volume_gb volume_gb
# TYPE openstack_cinder_volume_gb gauge
openstack_cinder_volume_gb{availability_zone="nova",bootable="false",id="6edbc2f4-1507-44f8-ac0d-eed1d2608d38",name="test-volume-attachments",server_id="f4fda93b-06e0-4743-8117-bc8bcecd651b",status="in-use",tenant_id="bab7d5c60cd041a0a36f7c4b6e1dd978",user_id="32779452fcd34ae1a53a797ac8a1e064",volume_type="lvmdriver-1"} 2
openstack_cinder_volume_gb{availability_zone="nova",bootable="true",id="173f7b48-c4c1-4e70-9acc-086b39073506",name="test-volume",server_id="",status="available",tenant_id="bab7d5c60cd041a0a36f7c4b6e1dd978",user_id="32779452fcd34ae1a53a797ac8a1e064",volume_type="lvmdriver-1"} 1
# HELP openstack_cinder_volume_status volume_status
# TYPE openstack_cinder_volume_status gauge
openstack_cinder_volume_status{bootable="false",id="6edbc2f4-1507-44f8-ac0d-eed1d2608d38",name="test-volume-attachments",server_id="f4fda93b-06e0-4743-8117-bc8bcecd651b",size="2",status="in-use",tenant_id="bab7d5c60cd041a0a36f7c4b6e1dd978",volume_type="lvmdriver-1"} 5
openstack_cinder_volume_status{bootable="true",id="173f7b48-c4c1-4e70-9acc-086b39073506",name="test-volume",server_id="",size="1",status="available",tenant_id="bab7d5c60cd041a0a36f7c4b6e1dd978",volume_type="lvmdriver-1"} 1
# HELP openstack_cinder_volume_status_counter volume_status_counter
# TYPE openstack_cinder_volume_status_counter gauge
openstack_cinder_volume_status_counter{status="attaching"} 0
openstack_cinder_volume_status_counter{status="available"} 1
openstack_cinder_volume_status_counter{status="awaiting-transfer"} 0
openstack_cinder_volume_status_counter{status="backing-up"} 0
openstack_cinder_volume_status_counter{status="creating"} 0
openstack_cinder_volume_status_counter{status="deleting"} 0
openstack_cinder_volume_status_counter{status="detaching"} 0
openstack_cinder_volume_status_counter{status="downloading"} 0
openstack_cinder_volume_status_counter{status="error"} 0
openstack_cinder_volume_status_counter{status="error_backing-up"} 0
openstack_cinder_volume_status_counter{status="error_deleting"} 0
openstack_cinder_volume_status_counter{status="error_extending"} 0
openstack_cinder_volume_status_counter{status="error_restoring"} 0
openstack_cinder_volume_status_counter{status="extending"} 0
openstack_cinder_volume_status_counter{status="in-use"} 1
openstack_cinder_volume_status_counter{status="maintenance"} 0
openstack_cinder_volume_status_counter{status="reserved"} 0
openstack_cinder_volume_status_counter{status="restoring-backup"} 0
openstack_cinder_volume_status_counter{status="retyping"} 0
openstack_cinder_volume_status_counter{status="uploading"} 0
# HELP openstack_cinder_volumes volumes
# TYPE openstack_cinder_volumes gauge
openstack_cinder_volumes 2
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetAllVolumes)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
			ExpectError:     false,
		},
	}

	testutil.RunCollectorTests(t, tests, NewVolumesCollector)
}
