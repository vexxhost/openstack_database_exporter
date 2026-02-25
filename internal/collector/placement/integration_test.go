//go:build integration

package placement

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	itest "github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestIntegration_ResourcesCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := itest.NewMySQLContainer(t, "placement", "../../../sql/placement/schema.sql")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewResourcesCollector(db, logger)

		expected := `# HELP openstack_placement_up up
# TYPE openstack_placement_up gauge
openstack_placement_up 1
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with resources and allocations", func(t *testing.T) {
		itest.SeedSQL(t, db,
			// Resource providers (compute hosts)
			`INSERT INTO resource_providers (id, uuid, name, generation, root_provider_id) VALUES
			(1, 'rp-uuid-001', 'compute-001', 1, 1),
			(2, 'rp-uuid-002', 'compute-002', 1, 2)`,
			// Resource classes
			`INSERT INTO resource_classes (id, name) VALUES
			(1, 'VCPU'),
			(2, 'MEMORY_MB'),
			(3, 'DISK_GB')`,
			// Inventories
			`INSERT INTO inventories (id, resource_provider_id, resource_class_id, total, reserved, min_unit, max_unit, step_size, allocation_ratio) VALUES
			(1, 1, 1, 64, 0, 1, 64, 1, 16.0000),
			(2, 1, 2, 131072, 512, 1, 131072, 1, 1.5000),
			(3, 1, 3, 1000, 0, 1, 1000, 1, 1.0000),
			(4, 2, 1, 32, 0, 1, 32, 1, 16.0000),
			(5, 2, 2, 65536, 256, 1, 65536, 1, 1.5000)`,
			// Allocations
			`INSERT INTO allocations (id, resource_provider_id, consumer_id, resource_class_id, used) VALUES
			(1, 1, 'inst-001', 1, 4),
			(2, 1, 'inst-001', 2, 8192),
			(3, 1, 'inst-001', 3, 40),
			(4, 1, 'inst-002', 1, 2),
			(5, 2, 'inst-003', 1, 8),
			(6, 2, 'inst-003', 2, 16384)`,
		)

		collector := NewResourcesCollector(db, logger)

		// 5 inventories × 4 metrics (total, ratio, reserved, usage) = 20
		// + 1 up = 21
		count := testutil.CollectAndCount(collector)
		if count != 21 {
			t.Fatalf("expected 21 metrics, got %d", count)
		}

		// Verify specific resource values
		expected := `# HELP openstack_placement_resource_total resource_total
# TYPE openstack_placement_resource_total gauge
openstack_placement_resource_total{hostname="compute-001",resourcetype="DISK_GB"} 1000
openstack_placement_resource_total{hostname="compute-001",resourcetype="MEMORY_MB"} 131072
openstack_placement_resource_total{hostname="compute-001",resourcetype="VCPU"} 64
openstack_placement_resource_total{hostname="compute-002",resourcetype="MEMORY_MB"} 65536
openstack_placement_resource_total{hostname="compute-002",resourcetype="VCPU"} 32
# HELP openstack_placement_resource_usage resource_usage
# TYPE openstack_placement_resource_usage gauge
openstack_placement_resource_usage{hostname="compute-001",resourcetype="DISK_GB"} 40
openstack_placement_resource_usage{hostname="compute-001",resourcetype="MEMORY_MB"} 8192
openstack_placement_resource_usage{hostname="compute-001",resourcetype="VCPU"} 6
openstack_placement_resource_usage{hostname="compute-002",resourcetype="MEMORY_MB"} 16384
openstack_placement_resource_usage{hostname="compute-002",resourcetype="VCPU"} 8
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
			"openstack_placement_resource_total",
			"openstack_placement_resource_usage",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
