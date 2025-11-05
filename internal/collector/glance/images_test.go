package glance

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	glancedb "github.com/vexxhost/openstack_database_exporter/internal/db/glance"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestImagesCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with images",
			SetupMock: func(mock sqlmock.Sqlmock) {
				createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
				updatedAt := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

				rows := sqlmock.NewRows([]string{
					"id", "name", "size", "status", "owner", "visibility",
					"disk_format", "container_format", "checksum", "created_at",
					"updated_at", "min_disk", "min_ram", "protected", "virtual_size",
					"os_hidden", "os_hash_algo", "os_hash_value",
				}).AddRow(
					"781b3762-9469-4cec-b58d-3349e5de4e9c", "F17-x86_64-cfntools", 476704768, "active", "5ef70662f8b34079a6eddb8da9d75fe8", "public",
					"qcow2", "bare", "1234567890abcdef", createdAt, updatedAt, 1, 512, false, nil,
					false, nil, nil,
				).AddRow(
					"1bea47ed-f6a9-463b-b423-14b9cca9ad27", "cirros-0.3.2-x86_64-disk", 13167616, "active", "5ef70662f8b34079a6eddb8da9d75fe8", "public",
					"qcow2", "bare", "abcdef1234567890", createdAt, updatedAt, 0, 64, false, nil,
					false, nil, nil,
				)

				mock.ExpectQuery(regexp.QuoteMeta(glancedb.GetAllImages)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_glance_image_bytes image_bytes
# TYPE openstack_glance_image_bytes gauge
openstack_glance_image_bytes{id="1bea47ed-f6a9-463b-b423-14b9cca9ad27",name="cirros-0.3.2-x86_64-disk",tenant_id="5ef70662f8b34079a6eddb8da9d75fe8"} 1.3167616e+07
openstack_glance_image_bytes{id="781b3762-9469-4cec-b58d-3349e5de4e9c",name="F17-x86_64-cfntools",tenant_id="5ef70662f8b34079a6eddb8da9d75fe8"} 4.76704768e+08
# HELP openstack_glance_image_created_at image_created_at
# TYPE openstack_glance_image_created_at gauge
openstack_glance_image_created_at{hidden="false",id="1bea47ed-f6a9-463b-b423-14b9cca9ad27",name="cirros-0.3.2-x86_64-disk",status="active",tenant_id="5ef70662f8b34079a6eddb8da9d75fe8",visibility="public"} 1.6725312e+09
openstack_glance_image_created_at{hidden="false",id="781b3762-9469-4cec-b58d-3349e5de4e9c",name="F17-x86_64-cfntools",status="active",tenant_id="5ef70662f8b34079a6eddb8da9d75fe8",visibility="public"} 1.6725312e+09
# HELP openstack_glance_images images
# TYPE openstack_glance_images gauge
openstack_glance_images 2
# HELP openstack_glance_up up
# TYPE openstack_glance_up gauge
openstack_glance_up 1
`,
		},
		{
			Name: "successful collection with no images",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "size", "status", "owner", "visibility",
					"disk_format", "container_format", "checksum", "created_at",
					"updated_at", "min_disk", "min_ram", "protected", "virtual_size",
					"os_hidden", "os_hash_algo", "os_hash_value",
				})

				mock.ExpectQuery(regexp.QuoteMeta(glancedb.GetAllImages)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_glance_images images
# TYPE openstack_glance_images gauge
openstack_glance_images 0
# HELP openstack_glance_up up
# TYPE openstack_glance_up gauge
openstack_glance_up 1
`,
		},
		{
			Name: "handles null values gracefully",
			SetupMock: func(mock sqlmock.Sqlmock) {
				createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

				rows := sqlmock.NewRows([]string{
					"id", "name", "size", "status", "owner", "visibility",
					"disk_format", "container_format", "checksum", "created_at",
					"updated_at", "min_disk", "min_ram", "protected", "virtual_size",
					"os_hidden", "os_hash_algo", "os_hash_value",
				}).AddRow(
					"image-with-nulls", nil, nil, "active", nil, "private",
					nil, nil, nil, createdAt, nil, 0, 0, false, nil,
					false, nil, nil,
				)

				mock.ExpectQuery(regexp.QuoteMeta(glancedb.GetAllImages)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_glance_image_bytes image_bytes
# TYPE openstack_glance_image_bytes gauge
openstack_glance_image_bytes{id="image-with-nulls",name="",tenant_id=""} 0
# HELP openstack_glance_image_created_at image_created_at
# TYPE openstack_glance_image_created_at gauge
openstack_glance_image_created_at{hidden="false",id="image-with-nulls",name="",status="active",tenant_id="",visibility="private"} 1.6725312e+09
# HELP openstack_glance_images images
# TYPE openstack_glance_images gauge
openstack_glance_images 1
# HELP openstack_glance_up up
# TYPE openstack_glance_up gauge
openstack_glance_up 1
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(glancedb.GetAllImages)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_glance_up up
# TYPE openstack_glance_up gauge
openstack_glance_up 0
`,
		},
	}

	testutil.RunCollectorTests(t, tests, NewImagesCollector)
}
