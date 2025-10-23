package octavia

import (
	"database/sql"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	octaviadb "github.com/vexxhost/openstack_database_exporter/internal/db/octavia"
)

func TestPoolCollector(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		expectedMetrics string
		expectError    bool
	}{
		{
			name: "successful collection with pools",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "project_id", "name", "protocol", "lb_algorithm",
					"operating_status", "load_balancer_id", "provisioning_status",
				}).AddRow(
					"pool1", "project1", "pool-one", "HTTP", "ROUND_ROBIN",
					"ONLINE", "lb1", "ACTIVE",
				).AddRow(
					"pool2", "project2", nil, "HTTPS", "LEAST_CONNECTIONS",
					"DEGRADED", "lb2", "ERROR",
				)

				mock.ExpectQuery(octaviadb.GetAllPools).WillReturnRows(rows)
			},
			expectedMetrics: `# HELP openstack_loadbalancer_pool_status pool_status
# TYPE openstack_loadbalancer_pool_status gauge
openstack_loadbalancer_pool_status{id="pool1",lb_algorithm="ROUND_ROBIN",loadbalancers="lb1",name="pool-one",operating_status="ONLINE",project_id="project1",protocol="HTTP",provisioning_status="ACTIVE"} 0
openstack_loadbalancer_pool_status{id="pool2",lb_algorithm="LEAST_CONNECTIONS",loadbalancers="lb2",name="",operating_status="DEGRADED",project_id="project2",protocol="HTTPS",provisioning_status="ERROR"} 2
# HELP openstack_loadbalancer_total_pools total_pools
# TYPE openstack_loadbalancer_total_pools gauge
openstack_loadbalancer_total_pools 2
`,
		},
		{
			name: "query error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(octaviadb.GetAllPools).WillReturnError(sql.ErrConnDone)
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
			collector := NewPoolCollector(db, logger)

			if tt.expectedMetrics != "" {
				err = testutil.CollectAndCompare(collector, strings.NewReader(tt.expectedMetrics))
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			} else {
				ch := make(chan prometheus.Metric, 10)
				collector.Collect(ch)
				close(ch)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
