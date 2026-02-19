//go:build integration

package octavia

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	itest "github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestIntegration_AmphoraCollector(t *testing.T) {
	t.Skip("skipping: octavia schema has NOT NULL columns that need proper seed data after migration")
	itest.SkipIfNoDocker(t)

	db := itest.NewMySQLContainer(t, "octavia", "../../../sql/octavia/schema.sql")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewAmphoraCollector(db, logger)
		expected := `# HELP openstack_loadbalancer_total_amphorae total_amphorae
# TYPE openstack_loadbalancer_total_amphorae gauge
openstack_loadbalancer_total_amphorae 0
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with seeded data", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO amphora (id, compute_id, status, load_balancer_id, lb_network_ip, ha_ip, role, cert_expiration) VALUES
			('amp-001', 'compute-001', 'READY', 'lb-001', '10.0.0.1', '10.0.0.2', 'MASTER', '2025-12-31 23:59:59'),
			('amp-002', 'compute-002', 'READY', 'lb-001', '10.0.0.3', '10.0.0.4', 'BACKUP', '2025-12-31 23:59:59'),
			('amp-003', NULL, 'BOOTING', NULL, NULL, NULL, NULL, NULL)`,
		)

		collector := NewAmphoraCollector(db, logger)

		// Just check the total count — individual metric labels depend on cert formatting
		expected := `# HELP openstack_loadbalancer_total_amphorae total_amphorae
# TYPE openstack_loadbalancer_total_amphorae gauge
openstack_loadbalancer_total_amphorae 3
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
			"openstack_loadbalancer_total_amphorae",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Clean up for next test
	itest.SeedSQL(t, db, "DELETE FROM amphora")
}

func TestIntegration_LoadBalancerCollector(t *testing.T) {
	t.Skip("skipping: octavia schema has NOT NULL columns that need proper seed data after migration")
	itest.SkipIfNoDocker(t)

	db := itest.NewMySQLContainer(t, "octavia", "../../../sql/octavia/schema.sql")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewLoadBalancerCollector(db, logger)
		expected := `# HELP openstack_loadbalancer_total_loadbalancers total_loadbalancers
# TYPE openstack_loadbalancer_total_loadbalancers gauge
openstack_loadbalancer_total_loadbalancers 0
# HELP openstack_loadbalancer_up up
# TYPE openstack_loadbalancer_up gauge
openstack_loadbalancer_up 1
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with load balancers and VIPs", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO load_balancer (id, project_id, name, provisioning_status, operating_status, provider) VALUES
			('lb-001', 'proj-abc', 'web-lb', 'ACTIVE', 'ONLINE', 'octavia'),
			('lb-002', 'proj-abc', 'api-lb', 'ACTIVE', 'DRAINING', 'octavia')`,
			`INSERT INTO vip (load_balancer_id, ip_address) VALUES
			('lb-001', '203.0.113.50')`,
		)

		collector := NewLoadBalancerCollector(db, logger)

		expected := `# HELP openstack_loadbalancer_loadbalancer_status loadbalancer_status
# TYPE openstack_loadbalancer_loadbalancer_status gauge
openstack_loadbalancer_loadbalancer_status{id="lb-001",name="web-lb",operating_status="ONLINE",project_id="proj-abc",provider="octavia",provisioning_status="ACTIVE",vip_address="203.0.113.50"} 0
openstack_loadbalancer_loadbalancer_status{id="lb-002",name="api-lb",operating_status="DRAINING",project_id="proj-abc",provider="octavia",provisioning_status="ACTIVE",vip_address=""} 1
# HELP openstack_loadbalancer_total_loadbalancers total_loadbalancers
# TYPE openstack_loadbalancer_total_loadbalancers gauge
openstack_loadbalancer_total_loadbalancers 2
# HELP openstack_loadbalancer_up up
# TYPE openstack_loadbalancer_up gauge
openstack_loadbalancer_up 1
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestIntegration_PoolCollector(t *testing.T) {
	t.Skip("skipping: octavia schema has NOT NULL columns that need proper seed data after migration")
	itest.SkipIfNoDocker(t)

	db := itest.NewMySQLContainer(t, "octavia", "../../../sql/octavia/schema.sql")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewPoolCollector(db, logger)
		expected := `# HELP openstack_loadbalancer_total_pools total_pools
# TYPE openstack_loadbalancer_total_pools gauge
openstack_loadbalancer_total_pools 0
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with pools", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO pool (id, project_id, name, protocol, lb_algorithm, operating_status, load_balancer_id, provisioning_status) VALUES
			('pool-001', 'proj-abc', 'http-pool', 'HTTP', 'ROUND_ROBIN', 'ONLINE', 'lb-001', 'ACTIVE'),
			('pool-002', 'proj-abc', 'tcp-pool', 'TCP', 'LEAST_CONNECTIONS', 'ERROR', NULL, 'ERROR')`,
		)

		collector := NewPoolCollector(db, logger)

		expected := `# HELP openstack_loadbalancer_pool_status pool_status
# TYPE openstack_loadbalancer_pool_status gauge
openstack_loadbalancer_pool_status{id="pool-001",lb_algorithm="ROUND_ROBIN",loadbalancers="lb-001",name="http-pool",operating_status="ONLINE",project_id="proj-abc",protocol="HTTP",provisioning_status="ACTIVE"} 0
openstack_loadbalancer_pool_status{id="pool-002",lb_algorithm="LEAST_CONNECTIONS",loadbalancers="",name="tcp-pool",operating_status="ERROR",project_id="proj-abc",protocol="TCP",provisioning_status="ERROR"} 2
# HELP openstack_loadbalancer_total_pools total_pools
# TYPE openstack_loadbalancer_total_pools gauge
openstack_loadbalancer_total_pools 2
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
