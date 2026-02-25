//go:build integration

package nova

import (
	"database/sql"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/expfmt"
	novadb "github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	novaapidb "github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
	itest "github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func novaDBS(t *testing.T) (*sql.DB, *sql.DB) {
	novaDB := itest.NewMySQLContainer(t, "nova", "../../../sql/nova/schema.sql", "../../../sql/nova/indexes.sql")
	novaAPIDB := itest.NewMySQLContainer(t, "nova-api", "../../../sql/nova_api/schema.sql", "../../../sql/nova_api/indexes.sql")
	return novaDB, novaAPIDB
}

func TestIntegration_ServerCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	novaDB, novaAPIDB := novaDBS(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewServerCollector(logger, novadb.New(novaDB), novaapidb.New(novaAPIDB))
		wrapper := &serverCollectorWrapper{collector}

		// Empty DB should emit total_vms=0 and availability_zones=0
		err := testutil.CollectAndCompare(wrapper, strings.NewReader(`# HELP openstack_nova_total_vms total_vms
# TYPE openstack_nova_total_vms gauge
openstack_nova_total_vms 0
# HELP openstack_nova_availability_zones availability_zones
# TYPE openstack_nova_availability_zones gauge
openstack_nova_availability_zones 0
`), "openstack_nova_total_vms", "openstack_nova_availability_zones")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("vm_state to API status mapping", func(t *testing.T) {
		// Seed a flavor so instance_type_id can be resolved
		itest.SeedSQL(t, novaAPIDB,
			`INSERT INTO flavors (id, flavorid, name, vcpus, memory_mb, swap, root_gb, ephemeral_gb) VALUES
			(1, 'flavor-uuid-1', 'm1.small', 1, 2048, 0, 20, 0)`,
		)

		// Seed instances with different vm_state values to test the mapping
		itest.SeedSQL(t, novaDB,
			`INSERT INTO instances (id, uuid, display_name, user_id, project_id, host,
				availability_zone, vm_state, power_state, task_state,
				memory_mb, vcpus, root_gb, ephemeral_gb,
				launched_at, terminated_at, instance_type_id, deleted) VALUES
			(1, 'uuid-active',  'srv-active',  'user-1', 'proj-1', 'compute-1', 'nova', 'active',   1, NULL, 2048, 1, 20, 0, NOW(), NULL, 1, 0),
			(2, 'uuid-stopped', 'srv-stopped', 'user-1', 'proj-1', 'compute-1', 'nova', 'stopped',  4, NULL, 2048, 1, 20, 0, NOW(), NULL, 1, 0),
			(3, 'uuid-build',   'srv-build',   'user-1', 'proj-1', 'compute-2', 'nova', 'building', 0, NULL, 2048, 1, 20, 0, NULL,  NULL, 1, 0),
			(4, 'uuid-resize',  'srv-resize',  'user-1', 'proj-1', 'compute-1', 'nova', 'resized',  1, NULL, 2048, 1, 20, 0, NOW(), NULL, 1, 0),
			(5, 'uuid-error',   'srv-error',   'user-1', 'proj-1', 'compute-2', 'nova', 'error',    1, NULL, 2048, 1, 20, 0, NOW(), NULL, 1, 0)`,
		)

		collector := NewServerCollector(logger, novadb.New(novaDB), novaapidb.New(novaAPIDB))
		wrapper := &serverCollectorWrapper{collector}

		// Collect all metrics and filter by server_status
		metrics := testutil.CollectAndCount(wrapper, "openstack_nova_server_status")
		if metrics != 5 {
			t.Fatalf("expected 5 server_status metrics, got %d", metrics)
		}

		// Verify total_vms
		err := testutil.CollectAndCompare(wrapper, strings.NewReader(`# HELP openstack_nova_total_vms total_vms
# TYPE openstack_nova_total_vms gauge
openstack_nova_total_vms 5
`), "openstack_nova_total_vms")
		if err != nil {
			t.Fatalf("unexpected total_vms error: %v", err)
		}

		// Gather all metrics and check status labels
		gathered, err := testutil.CollectAndFormat(wrapper, expfmt.TypeTextPlain, "openstack_nova_server_status")
		if err != nil {
			t.Fatalf("failed to collect: %v", err)
		}
		output := string(gathered)

		// Verify the key vm_state → API status translations
		statusChecks := map[string]string{
			"uuid-active":  `status="ACTIVE"`,
			"uuid-stopped": `status="SHUTOFF"`,
			"uuid-build":   `status="BUILD"`,
			"uuid-resize":  `status="VERIFY_RESIZE"`,
			"uuid-error":   `status="ERROR"`,
		}
		for id, expectedLabel := range statusChecks {
			if !strings.Contains(output, id) {
				t.Errorf("expected metric for instance %s not found in output", id)
				continue
			}
			// Find the line containing this UUID and check the status label
			for _, line := range strings.Split(output, "\n") {
				if strings.Contains(line, id) && !strings.Contains(line, expectedLabel) {
					t.Errorf("instance %s: expected %s but got line: %s", id, expectedLabel, line)
				}
			}
		}
	})

	t.Run("task_state overrides vm_state", func(t *testing.T) {
		// Seed instances with task_state that should override the default status
		itest.SeedSQL(t, novaDB,
			`INSERT INTO instances (id, uuid, display_name, user_id, project_id, host,
				availability_zone, vm_state, power_state, task_state,
				memory_mb, vcpus, root_gb, ephemeral_gb,
				launched_at, terminated_at, instance_type_id, deleted) VALUES
			(10, 'uuid-migrating', 'srv-migrating', 'user-1', 'proj-1', 'compute-1', 'nova', 'active',  1, 'migrating',       2048, 1, 20, 0, NOW(), NULL, 1, 0),
			(11, 'uuid-rebuilding','srv-rebuilding','user-1', 'proj-1', 'compute-1', 'nova', 'active',  1, 'rebuilding',      2048, 1, 20, 0, NOW(), NULL, 1, 0),
			(12, 'uuid-reverting', 'srv-reverting', 'user-1', 'proj-1', 'compute-1', 'nova', 'resized', 1, 'resize_reverting',2048, 1, 20, 0, NOW(), NULL, 1, 0)`,
		)

		collector := NewServerCollector(logger, novadb.New(novaDB), novaapidb.New(novaAPIDB))
		wrapper := &serverCollectorWrapper{collector}

		gathered, err := testutil.CollectAndFormat(wrapper, expfmt.TypeTextPlain, "openstack_nova_server_status")
		if err != nil {
			t.Fatalf("failed to collect: %v", err)
		}
		output := string(gathered)

		taskStateChecks := map[string]string{
			"uuid-migrating":  `status="MIGRATING"`,
			"uuid-rebuilding": `status="REBUILD"`,
			"uuid-reverting":  `status="REVERT_RESIZE"`,
		}
		for id, expectedLabel := range taskStateChecks {
			found := false
			for _, line := range strings.Split(output, "\n") {
				if strings.Contains(line, id) {
					found = true
					if !strings.Contains(line, expectedLabel) {
						t.Errorf("instance %s: expected %s but got line: %s", id, expectedLabel, line)
					}
				}
			}
			if !found {
				t.Errorf("expected metric for instance %s not found in output", id)
			}
		}
	})
}
