//go:build integration

package glance

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	itest "github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestIntegration_ImagesCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := itest.NewMySQLContainer(t, "glance", "../../../sql/glance/schema.sql")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewImagesCollector(db, logger)
		expected := `# HELP openstack_glance_images images
# TYPE openstack_glance_images gauge
openstack_glance_images 0
# HELP openstack_glance_up up
# TYPE openstack_glance_up gauge
openstack_glance_up 1
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
			"openstack_glance_images", "openstack_glance_up",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with images including deleted", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO images (id, name, size, status, created_at, deleted, min_disk, min_ram, visibility, os_hidden, owner) VALUES
			('img-001', 'ubuntu-22.04', 2147483648, 'active', '2024-01-15 10:30:00', 0, 0, 0, 'public', 0, 'admin-proj'),
			('img-002', 'cirros', 12345678, 'active', '2024-02-20 14:00:00', 0, 0, 512, 'shared', 0, 'admin-proj'),
			('img-003', 'deleted-image', 1000000, 'deleted', '2024-03-01 08:00:00', 1, 0, 0, 'private', 0, 'other-proj'),
			('img-004', NULL, NULL, 'queued', '2024-04-10 12:00:00', 0, 0, 0, 'community', 1, NULL)`,
		)

		collector := NewImagesCollector(db, logger)

		// deleted=1 images should be filtered out, so only 3 images
		expected := `# HELP openstack_glance_images images
# TYPE openstack_glance_images gauge
openstack_glance_images 3
# HELP openstack_glance_up up
# TYPE openstack_glance_up gauge
openstack_glance_up 1
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
			"openstack_glance_images", "openstack_glance_up",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify image_bytes metric count (should be 3 non-deleted images)
		count := testutil.CollectAndCount(collector, "openstack_glance_image_bytes")
		if count != 3 {
			t.Fatalf("expected 3 image_bytes metrics, got %d", count)
		}

		// Verify image_created_at metric count
		count = testutil.CollectAndCount(collector, "openstack_glance_image_created_at")
		if count != 3 {
			t.Fatalf("expected 3 image_created_at metrics, got %d", count)
		}
	})
}
