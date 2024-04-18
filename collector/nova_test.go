package collector

import (
	"strings"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

type NovaTestSuite struct {
	BaseOpenStackTestSuite
}

var novaExpectedUp = `
# HELP openstack_nova_server_task_state server_task_state
# TYPE openstack_nova_server_task_state gauge
openstack_nova_server_task_state{id="ec2917d8-cbd4-49b2-b204-f2c0a81cbe3b",task_state=""} 0
openstack_nova_server_task_state{id="f3e2e9b6-3b7d-4b1e-9e0d-0f6b3b3b1b1b",task_state="spawning"} 1
`

func (suite *NovaTestSuite) TestNovaCollector() {
	suite.mock.ExpectQuery("SELECT `uuid`,`task_state` FROM `instances` WHERE `deleted` = ?").WillReturnRows(
		sqlmock.NewRows([]string{"uuid", "task_state"}).
			AddRow("ec2917d8-cbd4-49b2-b204-f2c0a81cbe3b", nil).
			AddRow("f3e2e9b6-3b7d-4b1e-9e0d-0f6b3b3b1b1b", "spawning"),
	)

	collector := newNovaCollector(suite.logger, suite.db)

	err := testutil.CollectAndCompare(collector, strings.NewReader(novaExpectedUp))
	assert.NoError(suite.T(), err)
}
