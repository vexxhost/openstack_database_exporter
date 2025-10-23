package neutron

import (
	"log/slog"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHARouterAgentPortBindingCollector_Collect(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create a test logger
	logger := slog.Default()
	collector := NewHARouterAgentPortBindingCollector(db, logger)
	expectedSQL := regexp.QuoteMeta(`SELECT
    ha.router_id,
    ha.l3_agent_id,
    ha.state,
    a.host as agent_host,
    a.admin_state_up as agent_admin_state_up,
    a.heartbeat_timestamp as agent_heartbeat_timestamp
FROM
    ha_router_agent_port_bindings ha
    LEFT JOIN agents a ON ha.l3_agent_id = a.id`)

	t.Run("successful collection with active and standby routers", func(t *testing.T) {
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

		mock.ExpectQuery(expectedSQL).WillReturnRows(rows)
		expected := `
# HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
# TYPE openstack_neutron_l3_agent_of_router gauge
openstack_neutron_l3_agent_of_router{agent_admin_up="true",agent_alive="true",agent_host="dev-os-ctrl-02",ha_state="backup",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="f8a44de0-fc8e-45df-93c7-f79bf3b01c95"} 1
openstack_neutron_l3_agent_of_router{agent_admin_up="true",agent_alive="true",agent_host="dev-os-ctrl-02",ha_state="active",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f"} 1
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		assert.NoError(t, err)
	})

	t.Run("agent down when heartbeat too old", func(t *testing.T) {
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

		mock.ExpectQuery(expectedSQL).WillReturnRows(rows)
		expected := `
# HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
# TYPE openstack_neutron_l3_agent_of_router gauge
openstack_neutron_l3_agent_of_router{agent_admin_up="true",agent_alive="false",agent_host="dev-os-ctrl-02",ha_state="active",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f"} 0
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		assert.NoError(t, err)
	})

	t.Run("skip bindings with null agent_id", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"router_id",
			"l3_agent_id",
			"state",
			"agent_host",
			"agent_admin_state_up",
			"agent_heartbeat_timestamp",
		}).AddRow(
			"9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f",
			nil,
			"active",
			nil,
			nil,
			nil,
		)

		mock.ExpectQuery(expectedSQL).WillReturnRows(rows)
		expected := `
# HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
# TYPE openstack_neutron_l3_agent_of_router gauge
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		assert.NoError(t, err)
	})

	t.Run("handle query error gracefully", func(t *testing.T) {
		mock.ExpectQuery(expectedSQL).WillReturnError(assert.AnError)
		expected := `
# HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
# TYPE openstack_neutron_l3_agent_of_router gauge
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		assert.NoError(t, err)
	})

}
