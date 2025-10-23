package cinder

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestPoolsCollector(t *testing.T) {
	t.Skip("TODO: Not exposed via database.")

	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with pool capacity",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// TODO
			},
			ExpectedMetrics: `# HELP openstack_cinder_pool_capacity_free_gb pool_capacity_free_gb
# TYPE openstack_cinder_pool_capacity_free_gb gauge
openstack_cinder_pool_capacity_free_gb{name="i666testhost@FastPool01",vendor_name="EMC",volume_backend_name="VNX_Pool"} 636.316
# HELP openstack_cinder_pool_capacity_total_gb pool_capacity_total_gb
# TYPE openstack_cinder_pool_capacity_total_gb gauge
openstack_cinder_pool_capacity_total_gb{name="i666testhost@FastPool01",vendor_name="EMC",volume_backend_name="VNX_Pool"} 1692.429
`,
		},
		{
			Name: "empty pools",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// TODO
			},
			ExpectedMetrics: "",
		},
	}

	testutil.RunCollectorTests(t, tests, NewPoolsCollector)
}
