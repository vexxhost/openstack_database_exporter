package octavia

import (
	"database/sql"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	octaviadb "github.com/vexxhost/openstack_database_exporter/internal/db/octavia"
)

func TestLoadBalancerCollector_Collect(t *testing.T) {
	tests := []struct {
		name            string
		setupMock       func(sqlmock.Sqlmock)
		expectedMetrics string
		expectError     bool
	}{
		{
			name: "successful collection",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock load balancer query
				lbRows := sqlmock.NewRows([]string{
					"id", "project_id", "name", "provisioning_status",
					"operating_status", "provider", "vip_address",
				}).AddRow(
					"lb-1", "project-1", "test-lb", "ACTIVE",
					"ONLINE", "amphora", "192.168.1.100",
				).AddRow(
					"lb-2", "project-2", "test-lb-2", "ERROR",
					"OFFLINE", "ovn", "10.0.0.1",
				)

				mock.ExpectQuery(octaviadb.GetAllLoadBalancersWithVip).WillReturnRows(lbRows)
			},
			expectedMetrics: `# HELP openstack_loadbalancer_loadbalancer_status loadbalancer_status
# TYPE openstack_loadbalancer_loadbalancer_status gauge
openstack_loadbalancer_loadbalancer_status{id="lb-1",name="test-lb",operating_status="ONLINE",project_id="project-1",provider="amphora",provisioning_status="ACTIVE",vip_address="192.168.1.100"} 0
openstack_loadbalancer_loadbalancer_status{id="lb-2",name="test-lb-2",operating_status="OFFLINE",project_id="project-2",provider="ovn",provisioning_status="ERROR",vip_address="10.0.0.1"} 2
# HELP openstack_loadbalancer_total_loadbalancers total_loadbalancers
# TYPE openstack_loadbalancer_total_loadbalancers gauge
openstack_loadbalancer_total_loadbalancers 2
# HELP openstack_loadbalancer_up up
# TYPE openstack_loadbalancer_up gauge
openstack_loadbalancer_up 1
`,
		},
		{
			name: "handles query errors gracefully",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(octaviadb.GetAllLoadBalancersWithVip).WillReturnError(sql.ErrConnDone)
			},
			expectedMetrics: `# HELP openstack_loadbalancer_up up
# TYPE openstack_loadbalancer_up gauge
openstack_loadbalancer_up 0
`,
		},
		{
			name: "handles null values",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock load balancer query with nulls
				lbRows := sqlmock.NewRows([]string{
					"id", "project_id", "name", "provisioning_status",
					"operating_status", "provider", "vip_address",
				}).AddRow(
					"lb-2", nil, nil, "ERROR",
					"OFFLINE", nil, nil,
				)

				mock.ExpectQuery(octaviadb.GetAllLoadBalancersWithVip).WillReturnRows(lbRows)
			},
			expectedMetrics: `# HELP openstack_loadbalancer_loadbalancer_status loadbalancer_status
# TYPE openstack_loadbalancer_loadbalancer_status gauge
openstack_loadbalancer_loadbalancer_status{id="lb-2",name="",operating_status="OFFLINE",project_id="",provider="",provisioning_status="ERROR",vip_address=""} 2
# HELP openstack_loadbalancer_total_loadbalancers total_loadbalancers
# TYPE openstack_loadbalancer_total_loadbalancers gauge
openstack_loadbalancer_total_loadbalancers 1
# HELP openstack_loadbalancer_up up
# TYPE openstack_loadbalancer_up gauge
openstack_loadbalancer_up 1
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			collector := NewLoadBalancerCollector(db, logger)

			err = testutil.CollectAndCompare(collector, strings.NewReader(tt.expectedMetrics))
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
