package collector

import (
	"strings"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

type OctaviaTestSuite struct {
	BaseOpenStackTestSuite
}

var octaviaExpectedUp = `
# HELP openstack_loadbalancer_loadbalancer_status loadbalancer_status
# TYPE openstack_loadbalancer_loadbalancer_status gauge
openstack_loadbalancer_loadbalancer_status{id="607226db-27ef-4d41-ae89-f2a800e9c2db",name="best_load_balancer",operating_status="ONLINE",project_id="e3cd678b11784734bc366148aa37580e",provider="octavia",provisioning_status="ACTIVE",vip_address="203.0.113.50"} 0
`

func (suite *OctaviaTestSuite) TestOctaviaCollector() {
	suite.mock.ExpectQuery("SELECT `id`,`name`,`project_id`,`operating_status`,`provisioning_status`,`provider`,`VirtualIP`.`ip_address` AS `VirtualIP__ip_address` FROM `load_balancer` LEFT JOIN `vip` `VirtualIP` ON `load_balancer`.`id` = `VirtualIP`.`load_balancer_id`").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "project_id", "operating_status", "provisioning_status", "provider", "VirtualIP__ip_address"}).
			AddRow("607226db-27ef-4d41-ae89-f2a800e9c2db", "best_load_balancer", "e3cd678b11784734bc366148aa37580e", "ONLINE", "ACTIVE", "octavia", "203.0.113.50"),
	)

	collector := newOctaviaCollector(suite.logger, suite.db)

	err := testutil.CollectAndCompare(collector, strings.NewReader(octaviaExpectedUp))
	assert.NoError(suite.T(), err)
}
