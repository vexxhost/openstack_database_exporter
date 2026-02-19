//go:build integration

package neutron

import (
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/project"
	itest "github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func neutronDB(t *testing.T) *sql.DB {
	return itest.NewMySQLContainer(t, "neutron", "../../../sql/neutron/schema.sql")
}

func TestIntegration_HARouterAgentPortBindingCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := neutronDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewHARouterAgentPortBindingCollector(db, logger)
		expected := `# HELP openstack_neutron_up up
# TYPE openstack_neutron_up gauge
openstack_neutron_up 1
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with routers and alive agent", func(t *testing.T) {
		now := time.Now().Format("2006-01-02 15:04:05")

		itest.SeedSQL(t, db,
			fmt.Sprintf(`INSERT INTO agents (id, agent_type, binary, topic, host, admin_state_up, created_at, started_at, heartbeat_timestamp, configurations) VALUES
			('agent-001', 'L3 agent', 'neutron-l3-agent', 'l3_agent', 'ctrl-01', 1, '%s', '%s', '%s', '{}')`, now, now, now),
			`INSERT INTO ha_router_agent_port_bindings (port_id, router_id, l3_agent_id, state) VALUES
			('port-001', 'router-001', 'agent-001', 'active'),
			('port-002', 'router-002', 'agent-001', 'standby')`,
		)

		collector := NewHARouterAgentPortBindingCollector(db, logger)

		err := testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_neutron_up up
# TYPE openstack_neutron_up gauge
openstack_neutron_up 1
`), "openstack_neutron_up")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		count := testutil.CollectAndCount(collector, "openstack_neutron_l3_agent_of_router")
		if count != 2 {
			t.Fatalf("expected 2 l3_agent_of_router metrics, got %d", count)
		}
	})
}

func TestIntegration_FloatingIPCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := neutronDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewFloatingIPCollector(db, logger)
		expected := `# HELP openstack_neutron_floating_ips floating_ips
# TYPE openstack_neutron_floating_ips gauge
openstack_neutron_floating_ips 0
# HELP openstack_neutron_floating_ips_associated_not_active floating_ips_associated_not_active
# TYPE openstack_neutron_floating_ips_associated_not_active gauge
openstack_neutron_floating_ips_associated_not_active 0
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
			"openstack_neutron_floating_ips", "openstack_neutron_floating_ips_associated_not_active")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with floating IPs", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO floatingips (id, floating_ip_address, floating_network_id, floating_port_id, router_id, status, project_id, standard_attr_id) VALUES
			('fip-001', '203.0.113.10', 'ext-net-001', 'fport-001', 'router-001', 'ACTIVE', 'proj-001', 100),
			('fip-002', '203.0.113.11', 'ext-net-001', 'fport-002', 'router-001', 'DOWN', 'proj-001', 101),
			('fip-003', '203.0.113.12', 'ext-net-001', 'fport-003', NULL, 'DOWN', 'proj-001', 102)`,
		)

		collector := NewFloatingIPCollector(db, logger)

		err := testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_neutron_floating_ips floating_ips
# TYPE openstack_neutron_floating_ips gauge
openstack_neutron_floating_ips 3
# HELP openstack_neutron_floating_ips_associated_not_active floating_ips_associated_not_active
# TYPE openstack_neutron_floating_ips_associated_not_active gauge
openstack_neutron_floating_ips_associated_not_active 1
`), "openstack_neutron_floating_ips", "openstack_neutron_floating_ips_associated_not_active")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		count := testutil.CollectAndCount(collector, "openstack_neutron_floating_ip")
		if count != 3 {
			t.Fatalf("expected 3 floating_ip metrics, got %d", count)
		}
	})
}

func TestIntegration_RouterCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := neutronDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewRouterCollector(db, logger)
		expected := `# HELP openstack_neutron_routers routers
# TYPE openstack_neutron_routers gauge
openstack_neutron_routers 0
# HELP openstack_neutron_routers_not_active routers_not_active
# TYPE openstack_neutron_routers_not_active gauge
openstack_neutron_routers_not_active 0
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
			"openstack_neutron_routers", "openstack_neutron_routers_not_active")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with routers", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO routers (id, name, status, admin_state_up, project_id, gw_port_id, standard_attr_id) VALUES
			('rtr-001', 'main-router', 'ACTIVE', 1, 'proj-001', 'gwport-001', 200),
			('rtr-002', 'backup-router', 'ERROR', 1, 'proj-001', NULL, 201)`,
			`INSERT INTO ports (id, network_id, mac_address, admin_state_up, status, device_id, device_owner, standard_attr_id) VALUES
			('gwport-001', 'ext-net-001', 'fa:16:3e:00:00:01', 1, 'ACTIVE', 'rtr-001', 'network:router_gateway', 300)`,
		)

		collector := NewRouterCollector(db, logger)

		err := testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_neutron_routers routers
# TYPE openstack_neutron_routers gauge
openstack_neutron_routers 2
# HELP openstack_neutron_routers_not_active routers_not_active
# TYPE openstack_neutron_routers_not_active gauge
openstack_neutron_routers_not_active 1
`), "openstack_neutron_routers", "openstack_neutron_routers_not_active")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_neutron_router router
# TYPE openstack_neutron_router gauge
openstack_neutron_router{admin_state_up="true",external_network_id="ext-net-001",id="rtr-001",name="main-router",project_id="proj-001",status="ACTIVE"} 1
openstack_neutron_router{admin_state_up="true",external_network_id="",id="rtr-002",name="backup-router",project_id="proj-001",status="ERROR"} 1
`), "openstack_neutron_router")
		if err != nil {
			t.Fatalf("unexpected router error: %v", err)
		}
	})
}

func TestIntegration_NetworkCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := neutronDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewNetworkCollector(db, logger)
		expected := `# HELP openstack_neutron_networks networks
# TYPE openstack_neutron_networks gauge
openstack_neutron_networks 0
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "openstack_neutron_networks")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with networks", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO standardattributes (id, resource_type, created_at) VALUES (400, 'networks', NOW()), (401, 'networks', NOW())`,
			`INSERT INTO networks (id, name, status, project_id, standard_attr_id) VALUES
			('net-001', 'public-net', 'ACTIVE', 'proj-001', 400),
			('net-002', 'private-net', 'ACTIVE', 'proj-002', 401)`,
			`INSERT INTO networksegments (id, network_id, network_type, physical_network, segmentation_id, standard_attr_id) VALUES
			('seg-001', 'net-001', 'flat', 'physnet1', NULL, 402)`,
			`INSERT INTO standardattributes (id, resource_type, created_at) VALUES (402, 'segments', NOW())`,
			`INSERT INTO externalnetworks (network_id) VALUES ('net-001')`,
			`INSERT INTO networkrbacs (id, object_id, project_id, target_project, action) VALUES
			('rbac-001', 'net-001', 'proj-001', '*', 'access_as_external'),
			('rbac-002', 'net-002', 'proj-002', '*', 'access_as_shared')`,
			`INSERT INTO tags (standard_attr_id, tag) VALUES (400, 'env:prod')`,
		)

		collector := NewNetworkCollector(db, logger)

		err := testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_neutron_networks networks
# TYPE openstack_neutron_networks gauge
openstack_neutron_networks 2
`), "openstack_neutron_networks")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		count := testutil.CollectAndCount(collector, "openstack_neutron_network")
		if count != 2 {
			t.Fatalf("expected 2 network metrics, got %d", count)
		}
	})
}

func TestIntegration_PortCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := neutronDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewPortCollector(db, logger)
		expected := `# HELP openstack_neutron_ports ports
# TYPE openstack_neutron_ports gauge
openstack_neutron_ports 0
# HELP openstack_neutron_ports_lb_not_active ports_lb_not_active
# TYPE openstack_neutron_ports_lb_not_active gauge
openstack_neutron_ports_lb_not_active 0
# HELP openstack_neutron_ports_no_ips ports_no_ips
# TYPE openstack_neutron_ports_no_ips gauge
openstack_neutron_ports_no_ips 0
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
			"openstack_neutron_ports", "openstack_neutron_ports_lb_not_active", "openstack_neutron_ports_no_ips")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with ports including LB and no-IP", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO networks (id, name, status, project_id, standard_attr_id) VALUES
			('pnet-001', 'test-net', 'ACTIVE', 'proj-001', 500)
			ON DUPLICATE KEY UPDATE name=name`,
			`INSERT INTO standardattributes (id, resource_type, created_at) VALUES (500, 'networks', NOW())
			ON DUPLICATE KEY UPDATE id=id`,
			`INSERT INTO ports (id, network_id, mac_address, admin_state_up, status, device_id, device_owner, standard_attr_id, ip_allocation) VALUES
			('port-p01', 'pnet-001', 'fa:16:3e:aa:bb:01', 1, 'ACTIVE', 'dev-001', 'compute:nova', 501, 'immediate'),
			('port-p02', 'pnet-001', 'fa:16:3e:aa:bb:02', 1, 'DOWN', 'dev-002', 'neutron:LOADBALANCERV2', 502, 'immediate'),
			('port-p03', 'pnet-001', 'fa:16:3e:aa:bb:03', 1, 'ACTIVE', 'dev-003', 'network:dhcp', 503, 'none')`,
			`INSERT INTO ipallocations (port_id, ip_address, subnet_id, network_id) VALUES
			('port-p01', '10.0.0.1', 'sub-001', 'pnet-001')`,
		)

		collector := NewPortCollector(db, logger)

		err := testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_neutron_ports ports
# TYPE openstack_neutron_ports gauge
openstack_neutron_ports 3
# HELP openstack_neutron_ports_lb_not_active ports_lb_not_active
# TYPE openstack_neutron_ports_lb_not_active gauge
openstack_neutron_ports_lb_not_active 1
# HELP openstack_neutron_ports_no_ips ports_no_ips
# TYPE openstack_neutron_ports_no_ips gauge
openstack_neutron_ports_no_ips 0
`), "openstack_neutron_ports", "openstack_neutron_ports_lb_not_active", "openstack_neutron_ports_no_ips")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestIntegration_SecurityGroupCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := neutronDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewSecurityGroupCollector(db, logger)
		expected := `# HELP openstack_neutron_security_groups security_groups
# TYPE openstack_neutron_security_groups gauge
openstack_neutron_security_groups 0
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with security groups", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO securitygroups (id, name, project_id, standard_attr_id) VALUES
			('sg-001', 'default', 'proj-001', 600),
			('sg-002', 'web-sg', 'proj-001', 601)`,
		)

		collector := NewSecurityGroupCollector(db, logger)
		expected := `# HELP openstack_neutron_security_groups security_groups
# TYPE openstack_neutron_security_groups gauge
openstack_neutron_security_groups 2
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestIntegration_SubnetCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := neutronDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewSubnetCollector(db, logger)
		expected := `# HELP openstack_neutron_subnets subnets
# TYPE openstack_neutron_subnets gauge
openstack_neutron_subnets 0
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "openstack_neutron_subnets")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with subnets and IP availability", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO standardattributes (id, resource_type, created_at) VALUES
			(700, 'networks', NOW()), (701, 'subnets', NOW())
			ON DUPLICATE KEY UPDATE id=id`,
			`INSERT INTO networks (id, name, status, project_id, standard_attr_id) VALUES
			('snet-001', 'avail-net', 'ACTIVE', 'proj-001', 700)
			ON DUPLICATE KEY UPDATE name=name`,
			`INSERT INTO subnets (id, name, network_id, ip_version, cidr, gateway_ip, enable_dhcp, project_id, standard_attr_id) VALUES
			('sub-s01', 'test-subnet', 'snet-001', 4, '10.0.0.0/24', '10.0.0.1', 1, 'proj-001', 701)`,
			`INSERT INTO ipallocationpools (id, subnet_id, first_ip, last_ip) VALUES
			('pool-001', 'sub-s01', '10.0.0.2', '10.0.0.254')`,
			`INSERT INTO ipallocations (port_id, ip_address, subnet_id, network_id) VALUES
			('port-x01', '10.0.0.10', 'sub-s01', 'snet-001'),
			('port-x02', '10.0.0.11', 'sub-s01', 'snet-001')
			ON DUPLICATE KEY UPDATE port_id=port_id`,
		)

		collector := NewSubnetCollector(db, logger)

		err := testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_neutron_subnets subnets
# TYPE openstack_neutron_subnets gauge
openstack_neutron_subnets 1
`), "openstack_neutron_subnets")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify IP availability total: 10.0.0.2 to 10.0.0.254 = 253 IPs
		err = testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_neutron_network_ip_availabilities_total network_ip_availabilities_total
# TYPE openstack_neutron_network_ip_availabilities_total gauge
openstack_neutron_network_ip_availabilities_total{cidr="10.0.0.0/24",ip_version="4",network_id="snet-001",network_name="avail-net",project_id="proj-001",subnet_name="test-subnet"} 253
`), "openstack_neutron_network_ip_availabilities_total")
		if err != nil {
			t.Fatalf("unexpected ip_availabilities_total error: %v", err)
		}

		// Verify IP availability used: 2 allocations
		err = testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_neutron_network_ip_availabilities_used network_ip_availabilities_used
# TYPE openstack_neutron_network_ip_availabilities_used gauge
openstack_neutron_network_ip_availabilities_used{cidr="10.0.0.0/24",ip_version="4",network_id="snet-001",network_name="avail-net",project_id="proj-001",subnet_name="test-subnet"} 2
`), "openstack_neutron_network_ip_availabilities_used")
		if err != nil {
			t.Fatalf("unexpected ip_availabilities_used error: %v", err)
		}
	})
}

func TestIntegration_QuotaCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := neutronDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database no projects", func(t *testing.T) {
		resolver := project.NewResolver(logger, nil, 0)
		collector := NewQuotaCollector(db, logger, resolver)
		count := testutil.CollectAndCount(collector)
		if count != 0 {
			t.Fatalf("expected 0 metrics for empty quotas, got %d", count)
		}
	})

	t.Run("with quotas and resources", func(t *testing.T) {
		itest.SeedSQL(t, db,
			`INSERT INTO quotas (id, project_id, resource, `+"`limit`"+`) VALUES
			('q-001', 'proj-001', 'network', 200),
			('q-002', 'proj-001', 'router', 20)`,
			`INSERT INTO networks (id, name, status, project_id, standard_attr_id) VALUES
			('qnet-001', 'q-net-1', 'ACTIVE', 'proj-001', 800),
			('qnet-002', 'q-net-2', 'ACTIVE', 'proj-001', 801)
			ON DUPLICATE KEY UPDATE name=name`,
			`INSERT INTO standardattributes (id, resource_type, created_at) VALUES (800, 'networks', NOW()), (801, 'networks', NOW())
			ON DUPLICATE KEY UPDATE id=id`,
			`INSERT INTO routers (id, name, status, project_id, standard_attr_id) VALUES
			('qrtr-001', 'q-router', 'ACTIVE', 'proj-001', 802)
			ON DUPLICATE KEY UPDATE name=name`,
			`INSERT INTO standardattributes (id, resource_type, created_at) VALUES (802, 'routers', NOW())
			ON DUPLICATE KEY UPDATE id=id`,
		)

		resolver := project.NewResolver(logger, nil, 0)
		collector := NewQuotaCollector(db, logger, resolver)

		// proj-001 should have: network limit=200 (explicit), router limit=20 (explicit),
		// all others use defaults. 9 resources × 3 types (limit/reserved/used) = 27 metrics
		count := testutil.CollectAndCount(collector)
		if count != 27 {
			t.Fatalf("expected 27 metrics (9 resources × 3 types), got %d", count)
		}

		// Verify explicit network quota
		err := testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_neutron_quota_network quota_network
# TYPE openstack_neutron_quota_network gauge
openstack_neutron_quota_network{tenant="proj-001",type="limit"} 200
openstack_neutron_quota_network{tenant="proj-001",type="reserved"} 0
openstack_neutron_quota_network{tenant="proj-001",type="used"} 2
`), "openstack_neutron_quota_network")
		if err != nil {
			t.Fatalf("unexpected quota_network error: %v", err)
		}

		// Verify explicit router quota
		err = testutil.CollectAndCompare(collector, strings.NewReader(`# HELP openstack_neutron_quota_router quota_router
# TYPE openstack_neutron_quota_router gauge
openstack_neutron_quota_router{tenant="proj-001",type="limit"} 20
openstack_neutron_quota_router{tenant="proj-001",type="reserved"} 0
openstack_neutron_quota_router{tenant="proj-001",type="used"} 1
`), "openstack_neutron_quota_router")
		if err != nil {
			t.Fatalf("unexpected quota_router error: %v", err)
		}
	})
}
