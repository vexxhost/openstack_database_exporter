package nova

import (
	"database/sql"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	promtestutil "github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	novadb "github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	novaapidb "github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestServerCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with server data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "uuid", "display_name", "user_id", "project_id", "host",
					"availability_zone", "vm_state", "power_state", "task_state",
					"memory_mb", "vcpus", "root_gb", "ephemeral_gb",
					"launched_at", "terminated_at", "instance_type_id", "deleted",
				}).AddRow(
					1, "server-uuid-1", "test-server", "user-1", "project-1", "compute-1",
					"nova", "active", 1, nil,
					2048, 2, 20, 0,
					"2023-12-18 10:00:00", nil, 1, 0,
				).AddRow(
					2, "server-uuid-2", "test-server-2", "user-1", "project-1", "compute-2",
					"nova", "stopped", 4, nil,
					4096, 4, 40, 0,
					"2023-12-18 09:00:00", nil, 2, 0,
				)

				mock.ExpectQuery("SELECT (.+) FROM instances").WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "empty servers",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "uuid", "display_name", "user_id", "project_id", "host",
					"availability_zone", "vm_state", "power_state", "task_state",
					"memory_mb", "vcpus", "root_gb", "ephemeral_gb",
					"launched_at", "terminated_at", "instance_type_id", "deleted",
				})
				mock.ExpectQuery("SELECT (.+) FROM instances").WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM instances").WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		collector := NewServerCollector(logger, novadb.New(db), novaapidb.New(db))
		return &serverCollectorWrapper{collector}
	})
}

func TestServerCollectorTaskStateMetric(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	instanceColumns := []string{
		"id", "uuid", "display_name", "user_id", "project_id", "host",
		"availability_zone", "vm_state", "power_state", "task_state",
		"memory_mb", "vcpus", "root_gb", "ephemeral_gb",
		"launched_at", "terminated_at", "instance_type_id", "deleted",
	}
	instanceRows := sqlmock.NewRows(instanceColumns).
		AddRow(
			1, "ec2917d8-cbd4-49b2-b204-f2c0a81cbe3b", "server-without-task", "user-1", "project-1", "compute-1",
			"nova", "active", 1, nil,
			2048, 2, 20, 0,
			nil, nil, 1, 0,
		).
		AddRow(
			2, "f3e2e9b6-3b7d-4b1e-9e0d-0f6b3b3b1b1b", "server-with-task", "user-1", "project-1", "compute-2",
			"nova", "active", 1, "spawning",
			4096, 4, 40, 0,
			nil, nil, 2, 0,
		)
	mock.ExpectQuery("SELECT (.+) FROM instances").WillReturnRows(instanceRows)

	flavorColumns := []string{
		"id", "flavorid", "name", "vcpus", "memory_mb", "root_gb",
		"ephemeral_gb", "swap", "rxtx_factor", "disabled", "is_public",
	}
	mock.ExpectQuery("SELECT (.+) FROM flavors").WillReturnRows(sqlmock.NewRows(flavorColumns))

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	collector := NewServerCollector(logger, novadb.New(db), novaapidb.New(db))

	err = promtestutil.CollectAndCompare(
		&serverCollectorWrapper{collector},
		strings.NewReader(`# HELP openstack_nova_server_task_state server_task_state
# TYPE openstack_nova_server_task_state gauge
openstack_nova_server_task_state{id="ec2917d8-cbd4-49b2-b204-f2c0a81cbe3b",task_state=""} 0
openstack_nova_server_task_state{id="f3e2e9b6-3b7d-4b1e-9e0d-0f6b3b3b1b1b",task_state="spawning"} 1
`),
		"openstack_nova_server_task_state",
	)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

type serverCollectorWrapper struct {
	*ServerCollector
}

func (w *serverCollectorWrapper) Collect(ch chan<- prometheus.Metric) {
	_ = w.ServerCollector.Collect(ch)
}

func TestResolveServerStatus(t *testing.T) {
	tests := []struct {
		name      string
		vmState   string
		taskState string
		expected  string
	}{
		// Default vm_state mappings (no task_state)
		{"active", "active", "", "ACTIVE"},
		{"building", "building", "", "BUILD"},
		{"stopped", "stopped", "", "SHUTOFF"},
		{"resized", "resized", "", "VERIFY_RESIZE"},
		{"paused", "paused", "", "PAUSED"},
		{"suspended", "suspended", "", "SUSPENDED"},
		{"rescued", "rescued", "", "RESCUE"},
		{"error", "error", "", "ERROR"},
		{"deleted", "deleted", "", "DELETED"},
		{"soft-delete", "soft-delete", "", "SOFT_DELETED"},
		{"shelved", "shelved", "", "SHELVED"},
		{"shelved_offloaded", "shelved_offloaded", "", "SHELVED_OFFLOADED"},

		// task_state overrides for active vm_state
		{"active+rebuilding", "active", "rebuilding", "REBUILD"},
		{"active+rebuild_spawning", "active", "rebuild_spawning", "REBUILD"},
		{"active+migrating", "active", "migrating", "MIGRATING"},
		{"active+resize_prep", "active", "resize_prep", "RESIZE"},
		{"active+resize_migrating", "active", "resize_migrating", "RESIZE"},
		{"active+shelving", "active", "shelving", "SHELVED"},
		{"active+shelving_offloading", "active", "shelving_offloading", "SHELVED"},

		// task_state overrides for stopped vm_state
		{"stopped+resize_prep", "stopped", "resize_prep", "RESIZE"},
		{"stopped+rebuilding", "stopped", "rebuilding", "REBUILD"},

		// task_state overrides for resized vm_state
		{"resized+resize_reverting", "resized", "resize_reverting", "REVERT_RESIZE"},

		// task_state overrides for paused vm_state
		{"paused+migrating", "paused", "migrating", "MIGRATING"},

		// task_state that doesn't override (falls back to default)
		{"active+powering_off", "active", "powering_off", "ACTIVE"},
		{"stopped+powering_on", "stopped", "powering_on", "SHUTOFF"},

		// Unknown vm_state falls back to ToUpper
		{"unknown_state", "weird_state", "", "WEIRD_STATE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveServerStatus(tt.vmState, tt.taskState)
			assert.Equal(t, tt.expected, result)
		})
	}
}
