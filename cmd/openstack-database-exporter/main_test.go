//go:build integration

package main_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/prometheus/common/promslog"

	"github.com/vexxhost/openstack_database_exporter/internal/collector"
	itest "github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

// TestIntegration_E2E_FullExporter starts MySQL containers for all 11 services,
// seeds them with representative data, wires up the full exporter registry,
// serves /metrics via an httptest server, and validates the Prometheus
// exposition output.
func TestIntegration_E2E_FullExporter(t *testing.T) {
	itest.SkipIfNoDocker(t)

	sqlDir := "../../sql"

	// ── 1. Start containers ──────────────────────────────────────────────────
	cinderRes := itest.NewMySQLContainerWithURL(t, "cinder",
		sqlDir+"/cinder/schema.sql",
		sqlDir+"/cinder/indexes.sql",
	)
	glanceRes := itest.NewMySQLContainerWithURL(t, "glance", sqlDir+"/glance/schema.sql")
	keystoneRes := itest.NewMySQLContainerWithURL(t, "keystone", sqlDir+"/keystone/schema.sql")
	magnumRes := itest.NewMySQLContainerWithURL(t, "magnum", sqlDir+"/magnum/schema.sql")
	manilaRes := itest.NewMySQLContainerWithURL(t, "manila",
		sqlDir+"/manila/prereqs.sql",
		sqlDir+"/manila/schema.sql",
	)
	neutronRes := itest.NewMySQLContainerWithURL(t, "neutron", sqlDir+"/neutron/schema.sql")
	octaviaRes := itest.NewMySQLContainerWithURL(t, "octavia", sqlDir+"/octavia/schema.sql")
	placementRes := itest.NewMySQLContainerWithURL(t, "placement", sqlDir+"/placement/schema.sql")
	novaRes := itest.NewMySQLContainerWithURL(t, "nova",
		sqlDir+"/nova/schema.sql",
		sqlDir+"/nova/indexes.sql",
	)
	novaAPIRes := itest.NewMySQLContainerWithURL(t, "nova_api",
		sqlDir+"/nova_api/schema.sql",
		sqlDir+"/nova_api/indexes.sql",
	)
	heatRes := itest.NewMySQLContainerWithURL(t, "heat", sqlDir+"/heat/schema.sql")

	t.Log("All 11 MariaDB containers are up")

	// ── 2. Seed data ─────────────────────────────────────────────────────────

	// Cinder: volumes, snapshots, agents, quotas
	itest.SeedSQL(t, cinderRes.DB,
		`INSERT INTO volume_types (id, name, deleted) VALUES ('vtype-001', 'SSD', 0)`,
		`INSERT INTO volumes (id, display_name, size, status, availability_zone, bootable, project_id, user_id, volume_type_id, deleted) VALUES
		 ('vol-001', 'boot-vol', 40, 'in-use', 'nova', 1, 'proj-001', 'user-001', 'vtype-001', 0),
		 ('vol-002', 'data-vol', 100, 'available', 'nova', 0, 'proj-001', 'user-001', 'vtype-001', 0)`,
		`INSERT INTO volume_attachment (id, volume_id, instance_uuid, deleted) VALUES
		 ('att-001', 'vol-001', 'server-001', 0)`,
		`INSERT INTO snapshots (id, volume_id, volume_type_id, deleted, status) VALUES
		 ('snap-001', 'vol-001', 'vtype-001', 0, 'available'),
		 ('snap-002', 'vol-001', 'vtype-001', 0, 'creating')`,
		"INSERT INTO services (`host`, `binary`, `report_count`, `disabled`, `availability_zone`, `disabled_reason`, `updated_at`, `deleted`, `uuid`) VALUES"+
			" ('host-a@lvm', 'cinder-volume', 10, 0, 'nova', NULL, NOW(), 0, 'uuid-001')",
		`INSERT INTO quotas (project_id, resource, hard_limit, deleted) VALUES
		 ('proj-001', 'gigabytes', 1000, 0)`,
		`INSERT INTO quota_usages (project_id, resource, in_use, reserved, deleted) VALUES
		 ('proj-001', 'gigabytes', 250, 0, 0)`,
	)

	// Glance: images
	itest.SeedSQL(t, glanceRes.DB,
		`INSERT INTO images (id, name, size, status, created_at, deleted, min_disk, min_ram, visibility, os_hidden, owner) VALUES
		 ('img-001', 'ubuntu-22.04', 2147483648, 'active', '2024-01-15 10:30:00', 0, 0, 0, 'public', 0, 'admin-proj'),
		 ('img-002', 'cirros', 12345678, 'active', '2024-02-20 14:00:00', 0, 0, 512, 'shared', 0, 'admin-proj')`,
	)

	// Keystone: domains, projects, users, groups, regions
	itest.SeedSQL(t, keystoneRes.DB,
		`INSERT INTO project (id, name, enabled, domain_id, is_domain) VALUES
		 ('domain-001', 'TestDomain', 1, 'domain-001', 1)`,
		`INSERT INTO project (id, name, enabled, domain_id, parent_id, is_domain) VALUES
		 ('proj-001', 'test-project-1', 1, 'domain-001', 'domain-001', 0)`,
		`INSERT INTO project_tag (project_id, name) VALUES ('proj-001', 'env:prod')`,
		"INSERT INTO user (id, enabled, domain_id, created_at) VALUES ('user-001', 1, 'domain-001', NOW())",
		`INSERT INTO region (id, description) VALUES ('RegionOne', 'Primary region')`,
		"INSERT INTO `group` (id, domain_id, name) VALUES ('grp-001', 'domain-001', 'admins')",
	)

	// Magnum: clusters, nodegroups
	itest.SeedSQL(t, magnumRes.DB,
		`INSERT INTO cluster (uuid, name, stack_id, project_id, status) VALUES
		 ('clust-001', 'prod-cluster', 'stack-001', 'proj-001', 'CREATE_COMPLETE')`,
		`INSERT INTO nodegroup (uuid, name, cluster_id, project_id, role, node_count, is_default) VALUES
		 ('ng-001', 'master-prod', 'clust-001', 'proj-001', 'master', 3, 1),
		 ('ng-002', 'worker-prod', 'clust-001', 'proj-001', 'worker', 5, 1)`,
	)

	// Manila: shares, instances
	itest.SeedSQL(t, manilaRes.DB,
		`INSERT INTO availability_zones (id, name, deleted) VALUES ('az-001', 'nova', 'False')`,
		`INSERT INTO share_types (id, name, deleted) VALUES ('stype-001', 'default_share_type', 'False')`,
		`INSERT INTO shares (id, display_name, project_id, size, share_proto, deleted) VALUES
		 ('share-001', 'my-share', 'proj-001', 100, 'NFS', 'False')`,
		`INSERT INTO share_instances (id, share_id, status, share_type_id, availability_zone_id, deleted, cast_rules_to_readonly) VALUES
		 ('si-001', 'share-001', 'available', 'stype-001', 'az-001', 'False', 0)`,
	)

	// Neutron: agents, router bindings, routers, floating IPs, networks, subnets, ports, security groups, quotas
	now := time.Now().Format("2006-01-02 15:04:05")
	itest.SeedSQL(t, neutronRes.DB,
		fmt.Sprintf("INSERT INTO agents (id, agent_type, `binary`, topic, host, admin_state_up, created_at, started_at, heartbeat_timestamp, configurations) VALUES ('agent-001', 'L3 agent', 'neutron-l3-agent', 'l3_agent', 'ctrl-01', 1, '%s', '%s', '%s', '{}')", now, now, now),
		`INSERT INTO ha_router_agent_port_bindings (port_id, router_id, l3_agent_id, state) VALUES
		 ('port-001', 'router-001', 'agent-001', 'active')`,
		`INSERT INTO standardattributes (id, resource_type, created_at) VALUES
		 (1, 'networks', NOW()), (2, 'networks', NOW()), (3, 'subnets', NOW()),
		 (4, 'routers', NOW()), (5, 'ports', NOW()), (6, 'ports', NOW()),
		 (7, 'security_groups', NOW())`,
		`INSERT INTO networks (id, name, status, project_id, standard_attr_id) VALUES
		 ('net-001', 'public-net', 'ACTIVE', 'proj-001', 1),
		 ('net-002', 'private-net', 'ACTIVE', 'proj-001', 2)`,
		`INSERT INTO networksegments (id, network_id, network_type, standard_attr_id) VALUES
		 ('seg-001', 'net-001', 'flat', 8)`,
		`INSERT INTO standardattributes (id, resource_type, created_at) VALUES (8, 'segments', NOW())`,
		`INSERT INTO externalnetworks (network_id) VALUES ('net-001')`,
		`INSERT INTO subnets (id, name, network_id, ip_version, cidr, gateway_ip, enable_dhcp, project_id, standard_attr_id) VALUES
		 ('sub-001', 'public-subnet', 'net-001', 4, '10.0.0.0/24', '10.0.0.1', 1, 'proj-001', 3)`,
		`INSERT INTO ipallocationpools (id, subnet_id, first_ip, last_ip) VALUES
		 ('pool-001', 'sub-001', '10.0.0.2', '10.0.0.254')`,
		`INSERT INTO routers (id, name, status, admin_state_up, project_id, gw_port_id, standard_attr_id) VALUES
		 ('rtr-001', 'main-router', 'ACTIVE', 1, 'proj-001', 'gwport-001', 4)`,
		`INSERT INTO ports (id, network_id, mac_address, admin_state_up, status, device_id, device_owner, standard_attr_id, ip_allocation) VALUES
		 ('gwport-001', 'net-001', 'fa:16:3e:00:00:01', 1, 'ACTIVE', 'rtr-001', 'network:router_gateway', 5, 'immediate'),
		 ('port-n01', 'net-002', 'fa:16:3e:00:00:02', 1, 'ACTIVE', 'inst-001', 'compute:nova', 6, 'immediate')`,
		`INSERT INTO ipallocations (port_id, ip_address, subnet_id, network_id) VALUES
		 ('port-n01', '10.0.0.10', 'sub-001', 'net-001')`,
		`INSERT INTO floatingips (id, floating_ip_address, floating_network_id, floating_port_id, router_id, status, project_id, standard_attr_id) VALUES
		 ('fip-001', '203.0.113.10', 'net-001', 'fport-001', 'rtr-001', 'ACTIVE', 'proj-001', 9)`,
		`INSERT INTO standardattributes (id, resource_type, created_at) VALUES (9, 'floatingips', NOW())`,
		`INSERT INTO securitygroups (id, name, project_id, standard_attr_id) VALUES
		 ('sg-001', 'default', 'proj-001', 7)`,
		`INSERT INTO quotas (id, project_id, resource, `+"`limit`"+`) VALUES
		 ('q-001', 'proj-001', 'network', 200)`,
	)

	// Octavia: amphorae, load balancers, VIPs, pools
	itest.SeedSQL(t, octaviaRes.DB,
		`INSERT INTO amphora (id, compute_id, status, load_balancer_id, lb_network_ip, ha_ip, role, cert_expiration, cert_busy) VALUES
		 ('amp-001', 'compute-001', 'READY', 'lb-001', '10.0.0.1', '10.0.0.2', 'MASTER', '2025-12-31 23:59:59', 0)`,
		`INSERT INTO load_balancer (id, project_id, name, provisioning_status, operating_status, enabled, provider) VALUES
		 ('lb-001', 'proj-abc', 'web-lb', 'ACTIVE', 'ONLINE', 1, 'octavia')`,
		`INSERT INTO vip (load_balancer_id, ip_address) VALUES ('lb-001', '203.0.113.50')`,
		`INSERT INTO pool (id, project_id, name, protocol, lb_algorithm, operating_status, enabled, load_balancer_id, provisioning_status) VALUES
		 ('pool-001', 'proj-abc', 'http-pool', 'HTTP', 'ROUND_ROBIN', 'ONLINE', 1, 'lb-001', 'ACTIVE')`,
	)

	// Placement: resource providers, classes, inventories, allocations, projects, users, consumers
	itest.SeedSQL(t, placementRes.DB,
		`INSERT INTO resource_providers (id, uuid, name, generation, root_provider_id) VALUES
		 (1, 'rp-uuid-001', 'compute-001', 1, 1)`,
		`INSERT INTO resource_classes (id, name) VALUES (1, 'VCPU'), (2, 'MEMORY_MB')`,
		`INSERT INTO inventories (id, resource_provider_id, resource_class_id, total, reserved, min_unit, max_unit, step_size, allocation_ratio) VALUES
		 (1, 1, 1, 64, 0, 1, 64, 1, 16.0000),
		 (2, 1, 2, 131072, 512, 1, 131072, 1, 1.5000)`,
		`INSERT INTO projects (id, external_id) VALUES (1, 'proj-001')`,
		`INSERT INTO users (id, external_id) VALUES (1, 'user-001')`,
		`INSERT INTO consumers (id, uuid, project_id, user_id, generation) VALUES
		 (1, 'inst-001', 1, 1, 0)`,
		`INSERT INTO allocations (id, resource_provider_id, consumer_id, resource_class_id, used) VALUES
		 (1, 1, 'inst-001', 1, 4),
		 (2, 1, 'inst-001', 2, 8192)`,
	)

	// Nova: instances, services, compute_nodes
	itest.SeedSQL(t, novaRes.DB,
		`INSERT INTO instances (id, uuid, display_name, user_id, project_id, host, availability_zone, vm_state, power_state, memory_mb, vcpus, root_gb, ephemeral_gb, instance_type_id, deleted) VALUES
		 (1, 'inst-001', 'web-server', 'user-001', 'proj-001', 'compute-001', 'nova', 'active', 1, 2048, 2, 20, 0, 1, 0),
		 (2, 'inst-002', 'db-server', 'user-001', 'proj-001', 'compute-001', 'nova', 'active', 1, 4096, 4, 40, 0, 1, 0)`,
		`INSERT INTO services (id, host, `+"`binary`"+`, topic, report_count, disabled, deleted, disabled_reason, last_seen_up, forced_down, version, uuid) VALUES
		 (1, 'compute-001', 'nova-compute', 'compute', 100, 0, 0, NULL, NOW(), 0, 66, 'svc-uuid-001'),
		 (2, 'ctrl-001', 'nova-scheduler', 'scheduler', 200, 0, 0, NULL, NOW(), 0, 66, 'svc-uuid-002')`,
		`INSERT INTO compute_nodes (id, uuid, host, hypervisor_hostname, hypervisor_type, hypervisor_version, vcpus, vcpus_used, memory_mb, memory_mb_used, local_gb, local_gb_used, cpu_info, disk_available_least, free_ram_mb, free_disk_gb, current_workload, running_vms, cpu_allocation_ratio, ram_allocation_ratio, disk_allocation_ratio, deleted) VALUES
		 (1, 'cn-uuid-001', 'compute-001', 'compute-001.local', 'QEMU', 6001000, 64, 6, 131072, 6144, 1000, 60, '{}', 940, 124928, 940, 2, 2, 16.0, 1.5, 1.0, 0)`,
	)

	// Heat: stacks
	itest.SeedSQL(t, heatRes.DB,
		`INSERT INTO stack (id, name, action, status, tenant, raw_template_id, disable_rollback) VALUES
		 ('stack-001', 'web-app', 'CREATE', 'CREATE_COMPLETE', 'proj-001', 1, 0),
		 ('stack-002', 'db-app', 'CREATE', 'CREATE_COMPLETE', 'proj-001', 1, 0),
		 ('stack-003', 'failed-app', 'CREATE', 'CREATE_FAILED', 'proj-001', 1, 0)`,
	)

	// Nova API: flavors, quotas, aggregates, aggregate_hosts, quota_usages
	itest.SeedSQL(t, novaAPIRes.DB,
		`INSERT INTO flavors (id, name, memory_mb, vcpus, swap, flavorid, rxtx_factor, root_gb, ephemeral_gb, disabled, is_public) VALUES
		 (1, 'm1.small', 2048, 1, 0, '2', 1.0, 20, 0, 0, 1),
		 (2, 'm1.medium', 4096, 2, 0, '3', 1.0, 40, 0, 0, 1)`,
		`INSERT INTO quotas (id, project_id, resource, hard_limit) VALUES
		 (1, 'proj-001', 'cores', 40),
		 (2, 'proj-001', 'instances', 20),
		 (3, 'proj-001', 'ram', 102400)`,
		`INSERT INTO aggregates (id, uuid, name) VALUES
		 (1, 'agg-uuid-001', 'prod-aggregate')`,
		`INSERT INTO aggregate_hosts (id, host, aggregate_id) VALUES
		 (1, 'compute-001', 1)`,
		`INSERT INTO quota_usages (id, project_id, user_id, resource, in_use, reserved) VALUES
		 (1, 'proj-001', 'user-001', 'cores', 6, 0),
		 (2, 'proj-001', 'user-001', 'instances', 2, 0),
		 (3, 'proj-001', 'user-001', 'ram', 6144, 0)`,
	)

	t.Log("All 11 services seeded")

	// ── 3. Wire up the full exporter ─────────────────────────────────────────
	logger := promslog.New(&promslog.Config{})

	cfg := collector.Config{
		CinderDatabaseURL:    cinderRes.URL,
		GlanceDatabaseURL:    glanceRes.URL,
		KeystoneDatabaseURL:  keystoneRes.URL,
		MagnumDatabaseURL:    magnumRes.URL,
		ManilaDatabaseURL:    manilaRes.URL,
		NeutronDatabaseURL:   neutronRes.URL,
		OctaviaDatabaseURL:   octaviaRes.URL,
		PlacementDatabaseURL: placementRes.URL,
		NovaDatabaseURL:      novaRes.URL,
		NovaAPIDatabaseURL:   novaAPIRes.URL,
		HeatDatabaseURL:      heatRes.URL,
	}

	reg := collector.NewRegistry(cfg, logger)

	// ── 4. Serve metrics via httptest ────────────────────────────────────────
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("failed to GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	// ── 5. Parse Prometheus exposition format ────────────────────────────────
	parser := expfmt.NewTextParser(model.UTF8Validation)
	families, err := parser.TextToMetricFamilies(strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("failed to parse metrics: %v", err)
	}

	t.Logf("Parsed %d metric families from /metrics", len(families))

	// ── 6. Validate expected metric families ─────────────────────────────────

	// Every service must emit its _up metric = 1
	upMetrics := []string{
		"openstack_cinder_up",
		"openstack_glance_up",
		"openstack_identity_up",
		"openstack_container_infra_up",
		"openstack_sharev2_up",
		"openstack_neutron_up",
		"openstack_loadbalancer_up",
		"openstack_placement_up",
		"openstack_nova_up",
		"openstack_heat_up",
	}
	for _, name := range upMetrics {
		assertGaugeValue(t, families, name, 1)
	}

	// Cinder
	assertMetricExists(t, families, "openstack_cinder_agent_state")
	assertGaugeValue(t, families, "openstack_cinder_volumes", 2)
	assertGaugeValue(t, families, "openstack_cinder_snapshots", 2)
	assertMetricExists(t, families, "openstack_cinder_volume_gb")
	assertMetricExists(t, families, "openstack_cinder_volume_status")
	assertMetricExists(t, families, "openstack_cinder_volume_status_counter")
	assertMetricExists(t, families, "openstack_cinder_limits_volume_max_gb")
	assertMetricExists(t, families, "openstack_cinder_limits_volume_used_gb")

	// Glance
	assertGaugeValue(t, families, "openstack_glance_images", 2)
	assertMetricExists(t, families, "openstack_glance_image_bytes")
	assertMetricExists(t, families, "openstack_glance_image_created_at")

	// Keystone
	assertGaugeValue(t, families, "openstack_identity_domains", 1)
	assertGaugeValue(t, families, "openstack_identity_projects", 1)
	assertGaugeValue(t, families, "openstack_identity_users", 1)
	assertGaugeValue(t, families, "openstack_identity_groups", 1)
	assertGaugeValue(t, families, "openstack_identity_regions", 1)
	assertMetricExists(t, families, "openstack_identity_domain_info")
	assertMetricExists(t, families, "openstack_identity_project_info")

	// Magnum
	assertGaugeValue(t, families, "openstack_container_infra_total_clusters", 1)
	assertMetricExists(t, families, "openstack_container_infra_cluster_status")
	assertMetricExists(t, families, "openstack_container_infra_cluster_masters")
	assertMetricExists(t, families, "openstack_container_infra_cluster_nodes")

	// Manila
	assertGaugeValue(t, families, "openstack_sharev2_shares_counter", 1)
	assertMetricExists(t, families, "openstack_sharev2_share_gb")
	assertMetricExists(t, families, "openstack_sharev2_share_status")
	assertMetricExists(t, families, "openstack_sharev2_share_status_counter")

	// Neutron
	assertMetricExists(t, families, "openstack_neutron_agent_state")
	assertMetricExists(t, families, "openstack_neutron_l3_agent_of_router")
	assertGaugeValue(t, families, "openstack_neutron_floating_ips", 1)
	assertMetricExists(t, families, "openstack_neutron_floating_ip")
	assertGaugeValue(t, families, "openstack_neutron_networks", 2)
	assertMetricExists(t, families, "openstack_neutron_network")
	assertGaugeValue(t, families, "openstack_neutron_ports", 2)
	assertMetricExists(t, families, "openstack_neutron_port")
	assertGaugeValue(t, families, "openstack_neutron_routers", 1)
	assertMetricExists(t, families, "openstack_neutron_router")
	assertGaugeValue(t, families, "openstack_neutron_security_groups", 1)
	assertGaugeValue(t, families, "openstack_neutron_subnets", 1)
	assertMetricExists(t, families, "openstack_neutron_subnet")
	assertMetricExists(t, families, "openstack_neutron_network_ip_availabilities_total")
	assertMetricExists(t, families, "openstack_neutron_network_ip_availabilities_used")
	assertMetricExists(t, families, "openstack_neutron_quota_network")

	// Octavia
	assertGaugeValue(t, families, "openstack_loadbalancer_total_amphorae", 1)
	assertGaugeValue(t, families, "openstack_loadbalancer_total_loadbalancers", 1)
	assertGaugeValue(t, families, "openstack_loadbalancer_total_pools", 1)
	assertMetricExists(t, families, "openstack_loadbalancer_amphora_status")
	assertMetricExists(t, families, "openstack_loadbalancer_loadbalancer_status")
	assertMetricExists(t, families, "openstack_loadbalancer_pool_status")

	// Placement
	assertMetricExists(t, families, "openstack_placement_resource_total")
	assertMetricExists(t, families, "openstack_placement_resource_usage")
	assertMetricExists(t, families, "openstack_placement_resource_allocation_ratio")
	assertMetricExists(t, families, "openstack_placement_resource_reserved")

	// Nova
	assertMetricExists(t, families, "openstack_nova_agent_state")
	assertGaugeValue(t, families, "openstack_nova_flavors", 2)
	assertMetricExists(t, families, "openstack_nova_flavor")
	assertMetricExists(t, families, "openstack_nova_server_status")
	assertGaugeValue(t, families, "openstack_nova_total_vms", 2)
	assertMetricExists(t, families, "openstack_nova_server_local_gb")
	assertMetricExists(t, families, "openstack_nova_running_vms")
	assertMetricExists(t, families, "openstack_nova_vcpus_available")
	assertMetricExists(t, families, "openstack_nova_vcpus_used")
	assertMetricExists(t, families, "openstack_nova_memory_available_bytes")
	assertMetricExists(t, families, "openstack_nova_memory_used_bytes")
	assertMetricExists(t, families, "openstack_nova_quota_cores")
	assertMetricExists(t, families, "openstack_nova_limits_vcpus_max")

	// Heat
	assertMetricExists(t, families, "openstack_heat_stack_status_counter")

	// ── 7. Validate scrape timing ────────────────────────────────────────────
	start := time.Now()
	resp2, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("second scrape failed: %v", err)
	}
	resp2.Body.Close()
	elapsed := time.Since(start)
	t.Logf("Second scrape completed in %v", elapsed)
	if elapsed > 10*time.Second {
		t.Errorf("scrape took %v, expected < 10s", elapsed)
	}
}

// TestIntegration_E2E_PartialConfig tests that the exporter works when only
// some services are configured. Unconfigured services should be silently skipped.
func TestIntegration_E2E_PartialConfig(t *testing.T) {
	itest.SkipIfNoDocker(t)

	sqlDir := "../../sql"
	glanceRes := itest.NewMySQLContainerWithURL(t, "glance", sqlDir+"/glance/schema.sql")
	itest.SeedSQL(t, glanceRes.DB,
		`INSERT INTO images (id, name, size, status, created_at, deleted, min_disk, min_ram, visibility, os_hidden, owner) VALUES
		 ('img-001', 'cirros', 12345678, 'active', '2024-01-01 00:00:00', 0, 0, 0, 'public', 0, 'admin')`,
	)

	logger := promslog.New(&promslog.Config{})

	// Only Glance configured - all other URLs are empty
	cfg := collector.Config{
		GlanceDatabaseURL: glanceRes.URL,
	}
	reg := collector.NewRegistry(cfg, logger)

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("failed to GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	parser := expfmt.NewTextParser(model.UTF8Validation)
	families, err := parser.TextToMetricFamilies(strings.NewReader(string(bodyBytes)))
	if err != nil {
		t.Fatalf("failed to parse metrics: %v", err)
	}

	// Glance metrics should be present
	assertGaugeValue(t, families, "openstack_glance_up", 1)
	assertGaugeValue(t, families, "openstack_glance_images", 1)

	// Other services should NOT be present (empty URL = skip)
	for _, name := range []string{
		"openstack_cinder_up",
		"openstack_identity_up",
		"openstack_loadbalancer_up",
		"openstack_placement_up",
	} {
		if _, ok := families[name]; ok {
			t.Errorf("metric %s should not exist when service is not configured", name)
		}
	}
}

// ── P4.2 Scrape Simulation Tests ─────────────────────────────────────────────

// TestIntegration_E2E_ScrapeTiming seeds a realistic data volume into a single
// Glance container and verifies that repeated scrapes complete under 5 seconds.
func TestIntegration_E2E_ScrapeTiming(t *testing.T) {
	itest.SkipIfNoDocker(t)

	sqlDir := "../../sql"
	glanceRes := itest.NewMySQLContainerWithURL(t, "glance", sqlDir+"/glance/schema.sql")

	// Seed 500 images to simulate a moderate production workload.
	for i := 0; i < 500; i++ {
		itest.SeedSQL(t, glanceRes.DB, fmt.Sprintf(
			"INSERT INTO images (id, name, size, status, created_at, deleted, min_disk, min_ram, visibility, os_hidden, owner) VALUES ('img-%04d', 'image-%04d', %d, 'active', '2024-01-01 00:00:00', 0, 0, 0, 'public', 0, 'proj-001')",
			i, i, (i+1)*1024*1024,
		))
	}

	logger := promslog.New(&promslog.Config{})
	cfg := collector.Config{GlanceDatabaseURL: glanceRes.URL}
	reg := collector.NewRegistry(cfg, logger)

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Warm-up scrape
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("warm-up scrape failed: %v", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// Timed scrapes — run 5 consecutive scrapes
	for i := 0; i < 5; i++ {
		start := time.Now()
		resp, err := http.Get(ts.URL)
		if err != nil {
			t.Fatalf("scrape %d failed: %v", i, err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		elapsed := time.Since(start)
		t.Logf("Scrape %d: %v", i, elapsed)
		if elapsed > 5*time.Second {
			t.Errorf("scrape %d took %v, expected < 5s", i, elapsed)
		}
	}
}

// TestIntegration_E2E_ConcurrentScrapes verifies that the exporter handles
// multiple simultaneous scrape requests without data races or panics.
func TestIntegration_E2E_ConcurrentScrapes(t *testing.T) {
	itest.SkipIfNoDocker(t)

	sqlDir := "../../sql"
	glanceRes := itest.NewMySQLContainerWithURL(t, "glance", sqlDir+"/glance/schema.sql")
	cinderRes := itest.NewMySQLContainerWithURL(t, "cinder",
		sqlDir+"/cinder/schema.sql",
		sqlDir+"/cinder/indexes.sql",
	)

	// Seed some data
	itest.SeedSQL(t, glanceRes.DB,
		`INSERT INTO images (id, name, size, status, created_at, deleted, min_disk, min_ram, visibility, os_hidden, owner) VALUES
		 ('img-001', 'ubuntu', 2147483648, 'active', '2024-01-15 10:30:00', 0, 0, 0, 'public', 0, 'admin')`,
	)
	itest.SeedSQL(t, cinderRes.DB,
		`INSERT INTO volume_types (id, name, deleted) VALUES ('vt-001', 'SSD', 0)`,
		`INSERT INTO volumes (id, display_name, size, status, availability_zone, bootable, project_id, user_id, volume_type_id, deleted) VALUES
		 ('vol-001', 'test', 10, 'available', 'nova', 0, 'proj-001', 'user-001', 'vt-001', 0)`,
	)

	logger := promslog.New(&promslog.Config{})
	cfg := collector.Config{
		GlanceDatabaseURL: glanceRes.URL,
		CinderDatabaseURL: cinderRes.URL,
	}
	reg := collector.NewRegistry(cfg, logger)

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	const numWorkers = 10
	const scrapesPerWorker = 5
	var wg sync.WaitGroup
	errCh := make(chan error, numWorkers*scrapesPerWorker)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for s := 0; s < scrapesPerWorker; s++ {
				resp, err := http.Get(ts.URL)
				if err != nil {
					errCh <- fmt.Errorf("worker %d scrape %d: %v", workerID, s, err)
					continue
				}
				body, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					errCh <- fmt.Errorf("worker %d scrape %d read: %v", workerID, s, err)
					continue
				}
				if resp.StatusCode != http.StatusOK {
					errCh <- fmt.Errorf("worker %d scrape %d: status %d", workerID, s, resp.StatusCode)
					continue
				}
				// Verify we can parse the output
				parser := expfmt.NewTextParser(model.UTF8Validation)
				families, err := parser.TextToMetricFamilies(strings.NewReader(string(body)))
				if err != nil {
					errCh <- fmt.Errorf("worker %d scrape %d parse: %v", workerID, s, err)
					continue
				}
				if _, ok := families["openstack_glance_up"]; !ok {
					errCh <- fmt.Errorf("worker %d scrape %d: openstack_glance_up missing", workerID, s)
				}
				if _, ok := families["openstack_cinder_up"]; !ok {
					errCh <- fmt.Errorf("worker %d scrape %d: openstack_cinder_up missing", workerID, s)
				}
			}
		}(w)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		for _, err := range errs {
			t.Error(err)
		}
		t.Fatalf("%d errors in %d concurrent scrapes", len(errs), numWorkers*scrapesPerWorker)
	}
	t.Logf("Completed %d concurrent scrapes with 0 errors", numWorkers*scrapesPerWorker)
}

// TestIntegration_E2E_DBDownResilience verifies that when one service's database
// becomes unavailable, the exporter still serves metrics for the other services.
// The failed service should report _up=0.
func TestIntegration_E2E_DBDownResilience(t *testing.T) {
	itest.SkipIfNoDocker(t)

	sqlDir := "../../sql"

	// Start two containers: Glance (stays up) and Cinder (will be killed)
	glanceRes := itest.NewMySQLContainerWithURL(t, "glance", sqlDir+"/glance/schema.sql")
	cinderRes := itest.NewMySQLContainerWithURL(t, "cinder",
		sqlDir+"/cinder/schema.sql",
		sqlDir+"/cinder/indexes.sql",
	)

	// Seed data
	itest.SeedSQL(t, glanceRes.DB,
		`INSERT INTO images (id, name, size, status, created_at, deleted, min_disk, min_ram, visibility, os_hidden, owner) VALUES
		 ('img-001', 'ubuntu', 2147483648, 'active', '2024-01-15 10:30:00', 0, 0, 0, 'public', 0, 'admin')`,
	)
	itest.SeedSQL(t, cinderRes.DB,
		`INSERT INTO volume_types (id, name, deleted) VALUES ('vt-001', 'SSD', 0)`,
		`INSERT INTO volumes (id, display_name, size, status, availability_zone, bootable, project_id, user_id, volume_type_id, deleted) VALUES
		 ('vol-001', 'test', 10, 'in-use', 'nova', 1, 'proj-001', 'user-001', 'vt-001', 0)`,
	)

	logger := promslog.New(&promslog.Config{})
	cfg := collector.Config{
		GlanceDatabaseURL: glanceRes.URL,
		CinderDatabaseURL: cinderRes.URL,
	}

	// Register collectors while both DBs are alive
	reg := collector.NewRegistry(cfg, logger)

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// ── First scrape: both services should be up ─────────────────────────────
	families := scrapeAndParse(t, ts.URL)
	assertGaugeValue(t, families, "openstack_glance_up", 1)
	assertGaugeValue(t, families, "openstack_cinder_up", 1)
	assertGaugeValue(t, families, "openstack_glance_images", 1)
	assertGaugeValue(t, families, "openstack_cinder_volumes", 1)
	t.Log("First scrape: both services healthy")

	// ── Kill the Cinder database ─────────────────────────────────────────────
	// Terminate the Cinder container so that subsequent queries from the collector fail.
	// The collector has its own *sql.DB (from db.Connect), but the underlying MySQL
	// server is gone, so queries will error out.
	cinderRes.Terminate(context.Background())

	// Give connections a moment to realize they're dead
	time.Sleep(500 * time.Millisecond)

	// ── Second scrape: Glance should still work, Cinder should report _up=0 ──
	families = scrapeAndParse(t, ts.URL)

	// Glance: still healthy
	assertGaugeValue(t, families, "openstack_glance_up", 1)
	assertGaugeValue(t, families, "openstack_glance_images", 1)

	// Cinder: DB is down — each collector should report _up=0
	assertGaugeValue(t, families, "openstack_cinder_up", 0)

	t.Log("Second scrape: Glance healthy, Cinder correctly reports _up=0")
}

// scrapeAndParse performs a GET on the given URL and returns parsed metric families.
func scrapeAndParse(t *testing.T, url string) map[string]*dto.MetricFamily {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("scrape failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
	}

	parser := expfmt.NewTextParser(model.UTF8Validation)
	families, err := parser.TextToMetricFamilies(strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("failed to parse metrics: %v", err)
	}
	return families
}

// ── Assertion helpers ────────────────────────────────────────────────────────

func assertMetricExists(t *testing.T, families map[string]*dto.MetricFamily, name string) {
	t.Helper()
	if _, ok := families[name]; !ok {
		t.Errorf("expected metric %s to exist, but it was not found", name)
	}
}

func assertGaugeValue(t *testing.T, families map[string]*dto.MetricFamily, name string, expected float64) {
	t.Helper()
	mf, ok := families[name]
	if !ok {
		t.Errorf("expected metric %s to exist, but it was not found", name)
		return
	}

	metrics := mf.GetMetric()
	if len(metrics) == 0 {
		t.Errorf("metric %s has no samples", name)
		return
	}

	// Find a metric with the expected value
	for _, m := range metrics {
		if m.GetGauge() != nil && m.GetGauge().GetValue() == expected {
			return
		}
		if m.GetCounter() != nil && m.GetCounter().GetValue() == expected {
			return
		}
		if m.GetUntyped() != nil && m.GetUntyped().GetValue() == expected {
			return
		}
	}

	// Single metric: report mismatch
	if len(metrics) == 1 {
		var actual float64
		m := metrics[0]
		if m.GetGauge() != nil {
			actual = m.GetGauge().GetValue()
		} else if m.GetCounter() != nil {
			actual = m.GetCounter().GetValue()
		} else if m.GetUntyped() != nil {
			actual = m.GetUntyped().GetValue()
		}
		t.Errorf("metric %s: expected value %v, got %v", name, expected, actual)
	}
}
