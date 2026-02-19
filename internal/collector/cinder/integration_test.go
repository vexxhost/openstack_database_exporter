//go:build integration

package cinder

import (
	"database/sql"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/project"
	itest "github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func cinderDB(t *testing.T) *sql.DB {
	return itest.NewMySQLContainer(t, "cinder", "../../../sql/cinder/schema.sql", "../../../sql/cinder/indexes.sql")
}

func TestIntegration_VolumesCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := cinderDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewVolumesCollector(db, logger)

		// Should emit volumes=0, _up=1, and all 20 status_counter metrics (including reserved)
		count := testutil.CollectAndCount(collector)
		// 20 status counters + 1 volumes gauge + 1 up = 22
		if count != 22 {
			t.Fatalf("expected 22 metrics for empty volumes, got %d", count)
		}
	})

	t.Run("with volumes and attachments", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO volume_types (id, name, deleted) VALUES
			('vtype-001', 'SSD', 0),
			('vtype-002', 'HDD', 0)`,
			`INSERT INTO volumes (id, display_name, size, status, availability_zone, bootable, project_id, user_id, volume_type_id, deleted) VALUES
			('vol-001', 'boot-vol', 40, 'in-use', 'nova', 1, 'proj-001', 'user-001', 'vtype-001', 0),
			('vol-002', 'data-vol', 100, 'available', 'nova', 0, 'proj-001', 'user-001', 'vtype-002', 0),
			('vol-003', 'deleted-vol', 50, 'deleted', 'nova', 0, 'proj-002', 'user-002', 'vtype-001', 1)`,
			`INSERT INTO volume_attachment (id, volume_id, instance_uuid, deleted) VALUES
			('att-001', 'vol-001', 'server-001', 0)`,
		)

		collector := NewVolumesCollector(db, logger)

		// 2 active volumes × 2 metrics (volume_gb + volume_status) = 4
		// + 20 status counters + 1 volumes gauge + 1 up = 26
		count := testutil.CollectAndCount(collector)
		if count != 26 {
			t.Fatalf("expected 26 metrics, got %d", count)
		}

		// Verify volumes count
		expected := `# HELP openstack_cinder_volumes volumes
# TYPE openstack_cinder_volumes gauge
openstack_cinder_volumes 2
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "openstack_cinder_volumes")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify volume_status values are correct (reserved at index 2 shifts in-use to 5)
		err = testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_cinder_volume_status volume_status
# TYPE openstack_cinder_volume_status gauge
openstack_cinder_volume_status{bootable="true",id="vol-001",name="boot-vol",server_id="server-001",size="40",status="in-use",tenant_id="proj-001",volume_type="SSD"} 5
openstack_cinder_volume_status{bootable="false",id="vol-002",name="data-vol",server_id="",size="100",status="available",tenant_id="proj-001",volume_type="HDD"} 1
`), "openstack_cinder_volume_status")
		if err != nil {
			t.Fatalf("unexpected volume_status error: %v", err)
		}

		// Verify reserved status counter exists
		err = testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_cinder_volume_status_counter volume_status_counter
# TYPE openstack_cinder_volume_status_counter gauge
openstack_cinder_volume_status_counter{status="available"} 1
openstack_cinder_volume_status_counter{status="in-use"} 1
openstack_cinder_volume_status_counter{status="reserved"} 0
openstack_cinder_volume_status_counter{status="creating"} 0
openstack_cinder_volume_status_counter{status="attaching"} 0
openstack_cinder_volume_status_counter{status="detaching"} 0
openstack_cinder_volume_status_counter{status="maintenance"} 0
openstack_cinder_volume_status_counter{status="deleting"} 0
openstack_cinder_volume_status_counter{status="awaiting-transfer"} 0
openstack_cinder_volume_status_counter{status="error"} 0
openstack_cinder_volume_status_counter{status="error_deleting"} 0
openstack_cinder_volume_status_counter{status="backing-up"} 0
openstack_cinder_volume_status_counter{status="restoring-backup"} 0
openstack_cinder_volume_status_counter{status="error_backing-up"} 0
openstack_cinder_volume_status_counter{status="error_restoring"} 0
openstack_cinder_volume_status_counter{status="error_extending"} 0
openstack_cinder_volume_status_counter{status="downloading"} 0
openstack_cinder_volume_status_counter{status="uploading"} 0
openstack_cinder_volume_status_counter{status="retyping"} 0
openstack_cinder_volume_status_counter{status="extending"} 0
`), "openstack_cinder_volume_status_counter")
		if err != nil {
			t.Fatalf("unexpected volume_status_counter error: %v", err)
		}
	})
}

func TestIntegration_SnapshotsCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := cinderDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewSnapshotsCollector(db, logger)
		expected := `# HELP openstack_cinder_snapshots snapshots
# TYPE openstack_cinder_snapshots gauge
openstack_cinder_snapshots 0
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with snapshots, including deleted", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO volume_types (id, name, deleted) VALUES ('type-1', 'SSD', 0)`,
			`INSERT INTO volumes (id, volume_type_id, deleted) VALUES ('vol-001', 'type-1', 0)`,
			`INSERT INTO snapshots (id, volume_id, volume_type_id, deleted, status) VALUES
			('snap-001', 'vol-001', 'type-1', 0, 'available'),
			('snap-002', 'vol-001', 'type-1', 0, 'creating'),
			('snap-003', 'vol-001', 'type-1', 1, 'deleted')`,
		)

		collector := NewSnapshotsCollector(db, logger)
		expected := `# HELP openstack_cinder_snapshots snapshots
# TYPE openstack_cinder_snapshots gauge
openstack_cinder_snapshots 2
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestIntegration_AgentsCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := cinderDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewAgentsCollector(db, logger)
		// No services = no metrics emitted (no _up for agents)
		count := testutil.CollectAndCount(collector)
		if count != 0 {
			t.Fatalf("expected 0 metrics for empty agents, got %d", count)
		}
	})

	t.Run("with services", func(t *testing.T) {
		now := "NOW()"
		itest.SeedSQL(t, db,
			"INSERT INTO services (`host`, `binary`, `report_count`, `disabled`, `availability_zone`, `disabled_reason`, `updated_at`, `deleted`, `uuid`) VALUES"+
				" ('host-a@lvm', 'cinder-volume', 10, 0, 'nova', NULL, "+now+", 0, 'uuid-001'),"+
				" ('host-b', 'cinder-scheduler', 5, 1, 'nova', 'maintenance', "+now+", 0, 'uuid-002'),"+
				" ('host-c', 'cinder-backup', 1, 0, 'az-2', NULL, DATE_SUB(NOW(), INTERVAL 5 MINUTE), 0, 'uuid-003')",
		)

		collector := NewAgentsCollector(db, logger)

		// 3 agent_state metrics
		count := testutil.CollectAndCount(collector, "openstack_cinder_agent_state")
		if count != 3 {
			t.Fatalf("expected 3 agent_state metrics, got %d", count)
		}

		// Verify specific agent states: uuid-001 and uuid-002 should be up (updated NOW()),
		// uuid-003 should be down (updated 5 minutes ago)
		expected := `# HELP openstack_cinder_agent_state agent_state
# TYPE openstack_cinder_agent_state gauge
openstack_cinder_agent_state{adminState="enabled",disabledReason="",hostname="host-a@lvm",service="cinder-volume",uuid="uuid-001",zone="nova"} 1
openstack_cinder_agent_state{adminState="disabled",disabledReason="maintenance",hostname="host-b",service="cinder-scheduler",uuid="uuid-002",zone="nova"} 1
openstack_cinder_agent_state{adminState="enabled",disabledReason="",hostname="host-c",service="cinder-backup",uuid="uuid-003",zone="az-2"} 0
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "openstack_cinder_agent_state")
		if err != nil {
			t.Fatalf("unexpected agent_state error: %v", err)
		}
	})
}

func TestIntegration_LimitsCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := cinderDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		resolver := project.NewResolver(logger, nil, 0)
		collector := NewLimitsCollector(db, logger, resolver)
		count := testutil.CollectAndCount(collector)
		if count != 0 {
			t.Fatalf("expected 0 metrics for empty quotas, got %d", count)
		}
	})

	t.Run("with quotas, usage, and volume types", func(t *testing.T) {
		// Seed volume types for volume_type_quota_gigabytes
		itest.SeedSQL(t, db,
			`INSERT INTO volume_types (id, name, deleted) VALUES
			('vt-001', 'standard', 0),
			('vt-002', '__DEFAULT__', 0)`,
			`INSERT INTO quotas (project_id, resource, hard_limit, deleted) VALUES
			('proj-001', 'gigabytes', 1000, 0),
			('proj-001', 'backup_gigabytes', 500, 0),
			('proj-002', 'gigabytes', 2000, 0)`,
			`INSERT INTO quota_usages (project_id, resource, in_use, reserved, deleted) VALUES
			('proj-001', 'gigabytes', 250, 0, 0),
			('proj-001', 'backup_gigabytes', 50, 0, 0),
			('proj-002', 'gigabytes', 100, 0, 0)`,
		)

		// No keystone resolver — tenant name falls back to project ID
		resolver := project.NewResolver(logger, nil, 0)
		collector := NewLimitsCollector(db, logger, resolver)

		// Verify backup limits: proj-001 has explicit backup quota (500/50),
		// proj-002 has no backup quota so defaults apply (1000/0)
		err := testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_cinder_limits_backup_max_gb limits_backup_max_gb
# TYPE openstack_cinder_limits_backup_max_gb gauge
openstack_cinder_limits_backup_max_gb{tenant="proj-001",tenant_id="proj-001"} 500
openstack_cinder_limits_backup_max_gb{tenant="proj-002",tenant_id="proj-002"} 1000
# HELP openstack_cinder_limits_backup_used_gb limits_backup_used_gb
# TYPE openstack_cinder_limits_backup_used_gb gauge
openstack_cinder_limits_backup_used_gb{tenant="proj-001",tenant_id="proj-001"} 50
openstack_cinder_limits_backup_used_gb{tenant="proj-002",tenant_id="proj-002"} 0
`), "openstack_cinder_limits_backup_max_gb", "openstack_cinder_limits_backup_used_gb")
		if err != nil {
			t.Fatalf("unexpected backup limits error: %v", err)
		}

		// Verify volume limits
		err = testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_cinder_limits_volume_max_gb limits_volume_max_gb
# TYPE openstack_cinder_limits_volume_max_gb gauge
openstack_cinder_limits_volume_max_gb{tenant="proj-001",tenant_id="proj-001"} 1000
openstack_cinder_limits_volume_max_gb{tenant="proj-002",tenant_id="proj-002"} 2000
# HELP openstack_cinder_limits_volume_used_gb limits_volume_used_gb
# TYPE openstack_cinder_limits_volume_used_gb gauge
openstack_cinder_limits_volume_used_gb{tenant="proj-001",tenant_id="proj-001"} 250
openstack_cinder_limits_volume_used_gb{tenant="proj-002",tenant_id="proj-002"} 100
`), "openstack_cinder_limits_volume_max_gb", "openstack_cinder_limits_volume_used_gb")
		if err != nil {
			t.Fatalf("unexpected volume limits error: %v", err)
		}

		// Verify volume_type_quota_gigabytes: -1 for each project × type
		err = testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_cinder_volume_type_quota_gigabytes volume_type_quota_gigabytes
# TYPE openstack_cinder_volume_type_quota_gigabytes gauge
openstack_cinder_volume_type_quota_gigabytes{tenant="proj-001",tenant_id="proj-001",volume_type="__DEFAULT__"} -1
openstack_cinder_volume_type_quota_gigabytes{tenant="proj-001",tenant_id="proj-001",volume_type="standard"} -1
openstack_cinder_volume_type_quota_gigabytes{tenant="proj-002",tenant_id="proj-002",volume_type="__DEFAULT__"} -1
openstack_cinder_volume_type_quota_gigabytes{tenant="proj-002",tenant_id="proj-002",volume_type="standard"} -1
`), "openstack_cinder_volume_type_quota_gigabytes")
		if err != nil {
			t.Fatalf("unexpected volume_type_quota error: %v", err)
		}

		// Total metrics: 2 projects × 4 limit metrics + 2 projects × 2 volume_type_quotas = 12
		count := testutil.CollectAndCount(collector)
		if count != 12 {
			t.Fatalf("expected 12 limit metrics, got %d", count)
		}
	})
}
