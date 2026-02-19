//go:build integration

package manila

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	itest "github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestIntegration_SharesCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	// prereqs.sql creates the share_groups stub table required by shares FK
	db := itest.NewMySQLContainer(t, "manila",
		"../../../sql/manila/prereqs.sql",
		"../../../sql/manila/schema.sql",
	)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewSharesCollector(db, logger)

		// Should emit: up=1, shares_counter=0, and 19 status_counter metrics
		count := testutil.CollectAndCount(collector)
		// 1 up + 1 shares_counter + 19 share_status_counter = 21
		if count != 21 {
			t.Fatalf("expected 21 metrics for empty shares, got %d", count)
		}

		expected := `# HELP openstack_sharev2_shares_counter shares_counter
# TYPE openstack_sharev2_shares_counter gauge
openstack_sharev2_shares_counter 0
# HELP openstack_sharev2_up up
# TYPE openstack_sharev2_up gauge
openstack_sharev2_up 1
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
			"openstack_sharev2_shares_counter",
			"openstack_sharev2_up",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with shares", func(t *testing.T) {
		itest.SeedSQL(t, db,
			// Create availability zones
			`INSERT INTO availability_zones (id, name, deleted) VALUES
			('az-001', 'nova', 'False'),
			('az-002', 'az2', 'False')`,
			// Create share types
			`INSERT INTO share_types (id, name, deleted) VALUES
			('stype-001', 'default_share_type', 'False'),
			('stype-002', 'premium', 'False')`,
			// Create shares
			`INSERT INTO shares (id, display_name, project_id, size, share_proto, deleted) VALUES
			('share-001', 'my-share', 'proj-001', 100, 'NFS', 'False'),
			('share-002', 'data-share', 'proj-002', 500, 'CIFS', 'False'),
			('share-003', 'deleted-share', 'proj-001', 50, 'NFS', 'True')`,
			// Create share instances (status lives here)
			`INSERT INTO share_instances (id, share_id, status, share_type_id, availability_zone_id, deleted, cast_rules_to_readonly) VALUES
			('si-001', 'share-001', 'available', 'stype-001', 'az-001', 'False', 0),
			('si-002', 'share-002', 'creating', 'stype-002', 'az-002', 'False', 0)`,
		)

		collector := NewSharesCollector(db, logger)

		// 2 active shares × 2 (share_gb + share_status) = 4
		// + 1 up + 1 shares_counter + 19 status_counter = 25
		count := testutil.CollectAndCount(collector)
		if count != 25 {
			t.Fatalf("expected 25 metrics, got %d", count)
		}

		expected := `# HELP openstack_sharev2_shares_counter shares_counter
# TYPE openstack_sharev2_shares_counter gauge
openstack_sharev2_shares_counter 2
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
			"openstack_sharev2_shares_counter",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
