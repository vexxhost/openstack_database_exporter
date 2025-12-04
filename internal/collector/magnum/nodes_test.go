package magnum

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	magnumdb "github.com/vexxhost/openstack_database_exporter/internal/db/magnum"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestNodesCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with node data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"uuid", "name", "stack_id", "status", "project_id", "master_count", "node_count",
				}).AddRow(
					"273c39d5-fa17-4372-b6b1-93a572de2cef", "k8s", "31c1ee6c-081e-4f39-9f0f-f1d87a7defa1", "CREATE_FAILED", "0cbd49cbf76d405d9c86562e1d579bd3", int64(1), int64(1),
				)

				mock.ExpectQuery(regexp.QuoteMeta(magnumdb.GetClusterMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_container_infra_cluster_nodes cluster_nodes
# TYPE openstack_container_infra_cluster_nodes gauge
openstack_container_infra_cluster_nodes{master_count="1",name="k8s",project_id="0cbd49cbf76d405d9c86562e1d579bd3",stack_id="31c1ee6c-081e-4f39-9f0f-f1d87a7defa1",status="CREATE_FAILED",uuid="273c39d5-fa17-4372-b6b1-93a572de2cef"} 1
`,
		},
		{
			Name: "successful collection with multiple clusters",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"uuid", "name", "stack_id", "status", "project_id", "master_count", "node_count",
				}).AddRow(
					"cluster-1", "test-cluster-1", "stack-1", "CREATE_COMPLETE", "project-1", int64(3), int64(5),
				).AddRow(
					"cluster-2", "test-cluster-2", "stack-2", "UPDATE_IN_PROGRESS", "project-2", int64(1), int64(2),
				)

				mock.ExpectQuery(regexp.QuoteMeta(magnumdb.GetClusterMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_container_infra_cluster_nodes cluster_nodes
# TYPE openstack_container_infra_cluster_nodes gauge
openstack_container_infra_cluster_nodes{master_count="3",name="test-cluster-1",project_id="project-1",stack_id="stack-1",status="CREATE_COMPLETE",uuid="cluster-1"} 5
openstack_container_infra_cluster_nodes{master_count="1",name="test-cluster-2",project_id="project-2",stack_id="stack-2",status="UPDATE_IN_PROGRESS",uuid="cluster-2"} 2
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"uuid", "name", "stack_id", "status", "project_id", "master_count", "node_count",
				})

				mock.ExpectQuery(regexp.QuoteMeta(magnumdb.GetClusterMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "null values handling",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"uuid", "name", "stack_id", "status", "project_id", "master_count", "node_count",
				}).AddRow(
					nil, nil, "", "UNKNOWN_STATUS", nil, nil, nil,
				)

				mock.ExpectQuery(regexp.QuoteMeta(magnumdb.GetClusterMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_container_infra_cluster_nodes cluster_nodes
# TYPE openstack_container_infra_cluster_nodes gauge
openstack_container_infra_cluster_nodes{master_count="0",name="",project_id="",stack_id="",status="UNKNOWN_STATUS",uuid=""} 0
`,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(magnumdb.GetClusterMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		return &testNodesCollector{NewNodesCollector(db, logger)}
	})
}

type testNodesCollector struct {
	*NodesCollector
}

func (t *testNodesCollector) Collect(ch chan<- prometheus.Metric) {
	_ = t.NodesCollector.Collect(ch)
}
