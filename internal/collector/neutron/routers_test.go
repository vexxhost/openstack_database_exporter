package neutron

import (
	"database/sql"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

func TestHARouterAgentPortBindingCollector(t *testing.T) {
	tests := []struct {
		name            string
		setupMock       func(sqlmock.Sqlmock)
		expectedMetrics string
		expectError     bool
	}{
		{
			name: "successful collection with active and standby routers",
			setupMock: func(mock sqlmock.Sqlmock) {
				now := time.Now()
				rows := sqlmock.NewRows([]string{
					"router_id",
					"l3_agent_id",
					"state",
					"agent_host",
					"agent_admin_state_up",
					"agent_heartbeat_timestamp",
				}).AddRow(
					"9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f",
					"ddbf087c-e38f-4a73-bcb3-c38f2a719a03",
					"active",
					"dev-os-ctrl-02",
					1,
					now,
				).AddRow(
					"f8a44de0-fc8e-45df-93c7-f79bf3b01c95",
					"ddbf087c-e38f-4a73-bcb3-c38f2a719a03",
					"backup",
					"dev-os-ctrl-02",
					1,
					now,
				)

				mock.ExpectQuery(neutrondb.GetHARouterAgentPortBindingsWithAgents).WillReturnRows(rows)
			},
			expectedMetrics: `# HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
# TYPE openstack_neutron_l3_agent_of_router gauge
openstack_neutron_l3_agent_of_router{agent_admin_up="true",agent_alive="true",agent_host="dev-os-ctrl-02",ha_state="active",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f"} 1
openstack_neutron_l3_agent_of_router{agent_admin_up="true",agent_alive="true",agent_host="dev-os-ctrl-02",ha_state="backup",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="f8a44de0-fc8e-45df-93c7-f79bf3b01c95"} 1
`,
		},
		{
			name: "agent down when heartbeat too old",
			setupMock: func(mock sqlmock.Sqlmock) {
				oldTime := time.Now().Add(-2 * time.Minute)
				rows := sqlmock.NewRows([]string{
					"router_id",
					"l3_agent_id",
					"state",
					"agent_host",
					"agent_admin_state_up",
					"agent_heartbeat_timestamp",
				}).AddRow(
					"9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f",
					"ddbf087c-e38f-4a73-bcb3-c38f2a719a03",
					"active",
					"dev-os-ctrl-02",
					1,
					oldTime,
				)

				mock.ExpectQuery(neutrondb.GetHARouterAgentPortBindingsWithAgents).WillReturnRows(rows)
			},
			expectedMetrics: `# HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
# TYPE openstack_neutron_l3_agent_of_router gauge
openstack_neutron_l3_agent_of_router{agent_admin_up="true",agent_alive="false",agent_host="dev-os-ctrl-02",ha_state="active",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f"} 0
`,
		},
		{
			name: "agent admin down",
			setupMock: func(mock sqlmock.Sqlmock) {
				now := time.Now()
				rows := sqlmock.NewRows([]string{
					"router_id",
					"l3_agent_id",
					"state",
					"agent_host",
					"agent_admin_state_up",
					"agent_heartbeat_timestamp",
				}).AddRow(
					"9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f",
					"ddbf087c-e38f-4a73-bcb3-c38f2a719a03",
					"active",
					"dev-os-ctrl-02",
					0, // admin_state_up is false
					now,
				)

				mock.ExpectQuery(neutrondb.GetHARouterAgentPortBindingsWithAgents).WillReturnRows(rows)
			},
			expectedMetrics: `# HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
# TYPE openstack_neutron_l3_agent_of_router gauge
openstack_neutron_l3_agent_of_router{agent_admin_up="false",agent_alive="true",agent_host="dev-os-ctrl-02",ha_state="active",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f"} 1
`,
		},
		{
			name: "handles NULL values from LEFT JOIN",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"router_id",
					"l3_agent_id",
					"state",
					"agent_host",
					"agent_admin_state_up",
					"agent_heartbeat_timestamp",
				}).AddRow(
					"9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f",
					"ddbf087c-e38f-4a73-bcb3-c38f2a719a03",
					nil,
					nil,
					nil,
					nil,
				)

				mock.ExpectQuery(neutrondb.GetHARouterAgentPortBindingsWithAgents).WillReturnRows(rows)
			},
			expectedMetrics: `# HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
# TYPE openstack_neutron_l3_agent_of_router gauge
openstack_neutron_l3_agent_of_router{agent_admin_up="false",agent_alive="false",agent_host="",ha_state="",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f"} 0
`,
		},
		{
			name: "query error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(neutrondb.GetHARouterAgentPortBindingsWithAgents).WillReturnError(sql.ErrConnDone)
			},
			expectedMetrics: "",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			collector := NewHARouterAgentPortBindingCollector(db, logger)

			if tt.expectedMetrics != "" {
				err = testutil.CollectAndCompare(collector, strings.NewReader(tt.expectedMetrics))
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			} else {
				// Just collect to trigger any errors
				testutil.CollectAndLint(collector)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}