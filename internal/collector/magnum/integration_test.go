//go:build integration

package magnum

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	itest "github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestIntegration_ContainerInfraCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := itest.NewMySQLContainer(t, "magnum", "../../../sql/magnum/schema.sql")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewContainerInfraCollector(db, logger)

		expected := `# HELP openstack_container_infra_total_clusters total_clusters
# TYPE openstack_container_infra_total_clusters gauge
openstack_container_infra_total_clusters 0
# HELP openstack_container_infra_up up
# TYPE openstack_container_infra_up gauge
openstack_container_infra_up 1
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
			"openstack_container_infra_total_clusters",
			"openstack_container_infra_up",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with clusters and nodegroups", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO cluster (uuid, name, stack_id, project_id, status) VALUES
			('clust-001', 'prod-cluster', 'stack-001', 'proj-001', 'CREATE_COMPLETE'),
			('clust-002', 'dev-cluster', 'stack-002', 'proj-002', 'CREATE_IN_PROGRESS')`,
			`INSERT INTO nodegroup (uuid, name, cluster_id, project_id, role, node_count, is_default) VALUES
			('ng-001', 'master-prod', 'clust-001', 'proj-001', 'master', 3, 1),
			('ng-002', 'worker-prod', 'clust-001', 'proj-001', 'worker', 5, 1),
			('ng-003', 'master-dev', 'clust-002', 'proj-002', 'master', 1, 1),
			('ng-004', 'worker-dev', 'clust-002', 'proj-002', 'worker', 2, 1)`,
		)

		collector := NewContainerInfraCollector(db, logger)

		// 1 up + 1 total_clusters + 2 clusters × 3 (status + masters + nodes) = 8
		count := testutil.CollectAndCount(collector)
		if count != 8 {
			t.Fatalf("expected 8 metrics, got %d", count)
		}

		expected := `# HELP openstack_container_infra_total_clusters total_clusters
# TYPE openstack_container_infra_total_clusters gauge
openstack_container_infra_total_clusters 2
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
			"openstack_container_infra_total_clusters",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
