package collector

import (
	"strings"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

type NovaApiTestSuite struct {
	BaseOpenStackTestSuite
}

var novaApiExpectedUp = `
# HELP openstack_nova_api_build_request build_request
# TYPE openstack_nova_api_build_request gauge
openstack_nova_api_build_request{instance_uuid="f3e2e9b6-3b7d-4b1e-9e0d-0f6b3b3b1b1b",project_id="ec2917d8-cbd4-49b2-b204-f2c0a81cbe3b"} 1
openstack_nova_api_build_request{instance_uuid="894cacd1-a432-4093-a0e7-cd29503205da",project_id="107b88ab-f104-4ac5-8032-302e8a621d46"} 1
`

func (suite *NovaTestSuite) TestNovaApiCollector() {
	suite.mock.ExpectQuery("SELECT `project_id`,`instance_uuid` FROM `build_requests").WillReturnRows(
		sqlmock.NewRows([]string{"project_id", "instance_uuid"}).
			AddRow("ec2917d8-cbd4-49b2-b204-f2c0a81cbe3b", "f3e2e9b6-3b7d-4b1e-9e0d-0f6b3b3b1b1b").
			AddRow("107b88ab-f104-4ac5-8032-302e8a621d46", "894cacd1-a432-4093-a0e7-cd29503205da"),
	)

	collector := newNovaApiCollector(suite.logger, suite.db)

	err := testutil.CollectAndCompare(collector, strings.NewReader(novaApiExpectedUp))
	assert.NoError(suite.T(), err)
}
