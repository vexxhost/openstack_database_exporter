package magnum

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	magnumdb "github.com/vexxhost/openstack_database_exporter/internal/db/magnum"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestContainerInfraCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with magnum cluster",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"uuid", "name", "stack_id", "status", "project_id", "master_count", "node_count",
				}).AddRow(
					"273c39d5-fa17-4372-b6b1-93a572de2cef", "k8s", "31c1ee6c-081e-4f39-9f0f-f1d87a7defa1", "CREATE_FAILED", "0cbd49cbf76d405d9c86562e1d579bd3", int64(1), int64(1),
				)

				mock.ExpectQuery(regexp.QuoteMeta(magnumdb.GetClusterMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_container_infra_cluster_masters cluster_masters
# TYPE openstack_container_infra_cluster_masters gauge
openstack_container_infra_cluster_masters{name="k8s",node_count="1",project_id="0cbd49cbf76d405d9c86562e1d579bd3",stack_id="31c1ee6c-081e-4f39-9f0f-f1d87a7defa1",status="CREATE_FAILED",uuid="273c39d5-fa17-4372-b6b1-93a572de2cef"} 1
# HELP openstack_container_infra_cluster_nodes cluster_nodes
# TYPE openstack_container_infra_cluster_nodes gauge
openstack_container_infra_cluster_nodes{master_count="1",name="k8s",project_id="0cbd49cbf76d405d9c86562e1d579bd3",stack_id="31c1ee6c-081e-4f39-9f0f-f1d87a7defa1",status="CREATE_FAILED",uuid="273c39d5-fa17-4372-b6b1-93a572de2cef"} 1
# HELP openstack_container_infra_cluster_status cluster_status
# TYPE openstack_container_infra_cluster_status gauge
openstack_container_infra_cluster_status{master_count="1",name="k8s",node_count="1",project_id="0cbd49cbf76d405d9c86562e1d579bd3",stack_id="31c1ee6c-081e-4f39-9f0f-f1d87a7defa1",status="CREATE_FAILED",uuid="273c39d5-fa17-4372-b6b1-93a572de2cef"} 1
# HELP openstack_container_infra_total_clusters total_clusters
# TYPE openstack_container_infra_total_clusters gauge
openstack_container_infra_total_clusters 1
# HELP openstack_container_infra_up up
# TYPE openstack_container_infra_up gauge
openstack_container_infra_up 1
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
				).AddRow(
					"cluster-3", "test-cluster-3", "stack-3", "DELETE_FAILED", "project-1", int64(2), int64(3),
				)

				mock.ExpectQuery(regexp.QuoteMeta(magnumdb.GetClusterMetrics)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_container_infra_cluster_masters cluster_masters
# TYPE openstack_container_infra_cluster_masters gauge
openstack_container_infra_cluster_masters{name="test-cluster-1",node_count="5",project_id="project-1",stack_id="stack-1",status="CREATE_COMPLETE",uuid="cluster-1"} 3
openstack_container_infra_cluster_masters{name="test-cluster-2",node_count="2",project_id="project-2",stack_id="stack-2",status="UPDATE_IN_PROGRESS",uuid="cluster-2"} 1
openstack_container_infra_cluster_masters{name="test-cluster-3",node_count="3",project_id="project-1",stack_id="stack-3",status="DELETE_FAILED",uuid="cluster-3"} 2
# HELP openstack_container_infra_cluster_nodes cluster_nodes
# TYPE openstack_container_infra_cluster_nodes gauge
openstack_container_infra_cluster_nodes{master_count="3",name="test-cluster-1",project_id="project-1",stack_id="stack-1",status="CREATE_COMPLETE",uuid="cluster-1"} 5
openstack_container_infra_cluster_nodes{master_count="1",name="test-cluster-2",project_id="project-2",stack_id="stack-2",status="UPDATE_IN_PROGRESS",uuid="cluster-2"} 2
openstack_container_infra_cluster_nodes{master_count="2",name="test-cluster-3",project_id="project-1",stack_id="stack-3",status="DELETE_FAILED",uuid="cluster-3"} 3
# HELP openstack_container_infra_cluster_status cluster_status
# TYPE openstack_container_infra_cluster_status gauge
openstack_container_infra_cluster_status{master_count="3",name="test-cluster-1",node_count="5",project_id="project-1",stack_id="stack-1",status="CREATE_COMPLETE",uuid="cluster-1"} 0
openstack_container_infra_cluster_status{master_count="1",name="test-cluster-2",node_count="2",project_id="project-2",stack_id="stack-2",status="UPDATE_IN_PROGRESS",uuid="cluster-2"} 3
openstack_container_infra_cluster_status{master_count="2",name="test-cluster-3",node_count="3",project_id="project-1",stack_id="stack-3",status="DELETE_FAILED",uuid="cluster-3"} 7
# HELP openstack_container_infra_total_clusters total_clusters
# TYPE openstack_container_infra_total_clusters gauge
openstack_container_infra_total_clusters 3
# HELP openstack_container_infra_up up
# TYPE openstack_container_infra_up gauge
openstack_container_infra_up 1
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
			ExpectedMetrics: `# HELP openstack_container_infra_total_clusters total_clusters
# TYPE openstack_container_infra_total_clusters gauge
openstack_container_infra_total_clusters 0
# HELP openstack_container_infra_up up
# TYPE openstack_container_infra_up gauge
openstack_container_infra_up 1
`,
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
			ExpectedMetrics: `# HELP openstack_container_infra_cluster_masters cluster_masters
# TYPE openstack_container_infra_cluster_masters gauge
openstack_container_infra_cluster_masters{name="",node_count="0",project_id="",stack_id="",status="UNKNOWN_STATUS",uuid=""} 0
# HELP openstack_container_infra_cluster_nodes cluster_nodes
# TYPE openstack_container_infra_cluster_nodes gauge
openstack_container_infra_cluster_nodes{master_count="0",name="",project_id="",stack_id="",status="UNKNOWN_STATUS",uuid=""} 0
# HELP openstack_container_infra_cluster_status cluster_status
# TYPE openstack_container_infra_cluster_status gauge
openstack_container_infra_cluster_status{master_count="0",name="",node_count="0",project_id="",stack_id="",status="UNKNOWN_STATUS",uuid=""} -1
# HELP openstack_container_infra_total_clusters total_clusters
# TYPE openstack_container_infra_total_clusters gauge
openstack_container_infra_total_clusters 1
# HELP openstack_container_infra_up up
# TYPE openstack_container_infra_up gauge
openstack_container_infra_up 1
`,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(magnumdb.GetClusterMetrics)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: `# HELP openstack_container_infra_up up
# TYPE openstack_container_infra_up gauge
openstack_container_infra_up 1
`,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) *ContainerInfraCollector {
		return NewContainerInfraCollector(db, logger)
	})
}
