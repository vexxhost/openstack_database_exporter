package octavia

import (
	"database/sql"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	octaviadb "github.com/vexxhost/openstack_database_exporter/internal/db/octavia"
)

func TestAmphoraCollector(t *testing.T) {
	tests := []struct {
		name            string
		setupMock       func(sqlmock.Sqlmock)
		expectedMetrics string
		expectError     bool
	}{
		{
			name: "successful collection with amphorae",
			setupMock: func(mock sqlmock.Sqlmock) {
				certTime := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
				rows := sqlmock.NewRows([]string{
					"id", "compute_id", "status", "load_balancer_id",
					"lb_network_ip", "ha_ip", "role", "cert_expiration",
				}).AddRow(
					"amp1", "compute1", "ALLOCATED", "lb1",
					"10.0.0.5", "192.168.1.10", "MASTER", certTime,
				).AddRow(
					"amp2", "compute2", "ERROR", nil,
					nil, nil, nil, nil,
				)

				mock.ExpectQuery(octaviadb.GetAllAmphora).WillReturnRows(rows)
			},
			expectedMetrics: `# HELP openstack_loadbalancer_amphora_status amphora_status
# TYPE openstack_loadbalancer_amphora_status gauge
openstack_loadbalancer_amphora_status{cert_expiration="2024-12-31T23:59:59Z",compute_id="compute1",ha_ip="192.168.1.10",id="amp1",lb_network_ip="10.0.0.5",loadbalancer_id="lb1",role="MASTER",status="ALLOCATED"} 1
openstack_loadbalancer_amphora_status{cert_expiration="",compute_id="compute2",ha_ip="",id="amp2",lb_network_ip="",loadbalancer_id="",role="",status="ERROR"} 6
# HELP openstack_loadbalancer_total_amphorae total_amphorae
# TYPE openstack_loadbalancer_total_amphorae gauge
openstack_loadbalancer_total_amphorae 2
`,
		},
		{
			name: "query error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(octaviadb.GetAllAmphora).WillReturnError(sql.ErrConnDone)
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
			collector := NewAmphoraCollector(db, logger)

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
