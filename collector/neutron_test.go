package collector

import (
	"strings"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

type NeutronTestSuite struct {
	BaseOpenStackTestSuite
}

var neutronExpectedUp = `
# HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
# TYPE openstack_neutron_l3_agent_of_router gauge
openstack_neutron_l3_agent_of_router{agent_admin_up="true",agent_alive="true",agent_host="dev-os-ctrl-02",ha_state="",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f"} 1
openstack_neutron_l3_agent_of_router{agent_admin_up="true",agent_alive="true",agent_host="dev-os-ctrl-02",ha_state="",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="f8a44de0-fc8e-45df-93c7-f79bf3b01c95"} 1
`

func (suite *NeutronTestSuite) TestNeutronCollector() {
	suite.mock.ExpectQuery("SELECT `router_id`,`l3_agent_id`,`state`,`L3Agent`.`id` AS `L3Agent__id`,`L3Agent`.`host` AS `L3Agent__host`,`L3Agent`.`admin_state_up` AS `L3Agent__admin_state_up`,`L3Agent`.`heartbeat_timestamp` AS `L3Agent__heartbeat_timestamp` FROM `ha_router_agent_port_bindings` LEFT JOIN `agents` `L3Agent` ON `ha_router_agent_port_bindings`.`l3_agent_id` = `L3Agent`.`id`").WillReturnRows(
		sqlmock.NewRows([]string{"router_id", "l3_agent_id", "state", "L3Agent__id", "L3Agent__host", "L3Agent__admin_state_up", "L3Agent__heartbeat_timestamp"}).
			AddRow("9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f", "ddbf087c-e38f-4a73-bcb3-c38f2a719a03", "", "ddbf087c-e38f-4a73-bcb3-c38f2a719a03", "dev-os-ctrl-02", true, time.Now()).
			AddRow("f8a44de0-fc8e-45df-93c7-f79bf3b01c95", "ddbf087c-e38f-4a73-bcb3-c38f2a719a03", "", "ddbf087c-e38f-4a73-bcb3-c38f2a719a03", "dev-os-ctrl-02", true, time.Now()),
	)

	collector := newNeutronDatabaseCollector(suite.logger, suite.db)

	err := testutil.CollectAndCompare(collector, strings.NewReader(neutronExpectedUp))
	assert.NoError(suite.T(), err)
}
