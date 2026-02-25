-- name: GetAgents :many
SELECT
    a.id,
    a.agent_type,
    a.`binary` as service,
    a.host as hostname,
    CASE
        WHEN a.admin_state_up = 1 THEN 'enabled'
        ELSE 'disabled'
    END as admin_state,
    a.availability_zone as zone,
    CASE
        WHEN TIMESTAMPDIFF(SECOND, a.heartbeat_timestamp, NOW()) <= 75 THEN 1
        ELSE 0
    END as alive
FROM
    agents a;

-- name: GetHARouterAgentPortBindingsWithAgents :many
SELECT
    ha.router_id,
    ha.l3_agent_id,
    ha.state,
    a.host as agent_host,
    a.admin_state_up as agent_admin_state_up,
    a.heartbeat_timestamp as agent_heartbeat_timestamp
FROM
    ha_router_agent_port_bindings ha
    LEFT JOIN agents a ON ha.l3_agent_id = a.id;

-- name: GetRouters :many
SELECT
    r.id,
    r.name,
    r.status,
    r.admin_state_up,
    r.project_id,
    COALESCE(p.network_id, '') as external_network_id
FROM
    routers r
    LEFT JOIN ports p ON r.gw_port_id = p.id;

-- name: GetFloatingIPs :many
SELECT
    fip.id,
    fip.floating_ip_address,
    fip.floating_network_id,
    fip.project_id,
    fip.router_id,
    fip.status,
    fip.fixed_ip_address
FROM
    floatingips fip;

-- name: GetNetworks :many
SELECT
    n.id,
    n.name,
    n.project_id,
    n.status,
    ns.network_type as provider_network_type,
    ns.physical_network as provider_physical_network,
    COALESCE(CAST(ns.segmentation_id AS CHAR), '') as provider_segmentation_id,
    COALESCE(CAST(GROUP_CONCAT(DISTINCT s.id) AS CHAR), '') as subnets,
    CASE
        WHEN en.network_id IS NOT NULL THEN 1
        ELSE 0
    END AS is_external,
    CASE
        WHEN shared_rbacs.object_id IS NOT NULL THEN 1
        ELSE 0
    END AS is_shared,
    COALESCE(CAST(GROUP_CONCAT(DISTINCT t.tag) AS CHAR), '') as tags
FROM
    networks n
    LEFT JOIN networksegments ns ON n.id = ns.network_id
    LEFT JOIN subnets s ON n.id = s.network_id
    LEFT JOIN externalnetworks en ON n.id = en.network_id
    LEFT JOIN networkrbacs shared_rbacs ON n.id = shared_rbacs.object_id AND shared_rbacs.target_project = '*' AND shared_rbacs.action = 'access_as_shared'
    LEFT JOIN standardattributes sa ON n.standard_attr_id = sa.id
    LEFT JOIN tags t ON sa.id = t.standard_attr_id
GROUP BY
    n.id,
    n.name,
    n.project_id,
    n.status,
    ns.network_type,
    ns.physical_network,
    ns.segmentation_id,
    en.network_id,
    shared_rbacs.object_id;

-- name: GetSubnets :many
SELECT
    s.id,
    s.name,
    s.cidr,
    s.gateway_ip,
    s.network_id,
    s.project_id,
    s.enable_dhcp,
    COALESCE(CAST(GROUP_CONCAT(DISTINCT d.address) AS CHAR), '') as dns_nameservers,
    s.subnetpool_id,
    COALESCE(CAST(GROUP_CONCAT(DISTINCT t.tag) AS CHAR), '') as tags
FROM
    subnets s
    LEFT JOIN dnsnameservers d ON s.id = d.subnet_id
    LEFT JOIN standardattributes sa ON s.standard_attr_id = sa.id
    LEFT JOIN tags t ON sa.id = t.standard_attr_id
GROUP BY
    s.id,
    s.name,
    s.cidr,
    s.gateway_ip,
    s.network_id,
    s.project_id,
    s.enable_dhcp,
    s.subnetpool_id;

-- name: GetPorts :many
SELECT
    p.id,
    p.mac_address,
    p.device_owner,
    p.status,
    p.network_id,
    p.admin_state_up,
    p.ip_allocation,
    b.vif_type as binding_vif_type,
    COALESCE(CAST(GROUP_CONCAT(ia.ip_address ORDER BY ia.ip_address) AS CHAR), '') as fixed_ips
FROM
    ports p
    LEFT JOIN ml2_port_bindings b ON p.id = b.port_id
    LEFT JOIN ipallocations ia ON p.id = ia.port_id
GROUP BY
    p.id,
    p.mac_address,
    p.device_owner,
    p.status,
    p.network_id,
    p.admin_state_up,
    p.ip_allocation,
    b.vif_type;

-- name: GetSecurityGroupCount :one
SELECT
    CAST(COUNT(*) AS SIGNED) as cnt
FROM
    securitygroups;

-- name: GetNetworkIPAvailabilitiesUsed :many
SELECT
    s.id AS subnet_id,
    s.name AS subnet_name,
    s.cidr,
    s.ip_version,
    s.project_id,
    n.id AS network_id,
    n.name AS network_name,
    CAST(COUNT(ipa.ip_address) AS SIGNED) AS allocation_count
FROM subnets s
    LEFT JOIN ipallocations ipa ON ipa.subnet_id = s.id
    LEFT JOIN networks n ON s.network_id = n.id
GROUP BY s.id, n.id;

-- name: GetNetworkIPAvailabilitiesTotal :many
SELECT
    s.name AS subnet_name,
    n.name AS network_name,
    s.id   AS subnet_id,
    n.id   AS network_id,
    ap.first_ip,
    ap.last_ip,
    s.project_id,
    s.cidr,
    s.ip_version
FROM subnets s
JOIN networks n
    ON s.network_id = n.id
LEFT JOIN ipallocationpools ap
    ON s.id = ap.subnet_id
GROUP BY
    s.id,
    n.id,
    s.project_id,
    s.cidr,
    s.ip_version,
    s.name,
    n.name,
    ap.first_ip,
    ap.last_ip;

-- name: GetSubnetPools :many
SELECT
    sp.id,
    sp.ip_version,
    sp.max_prefixlen,
    sp.min_prefixlen,
    sp.default_prefixlen,
    sp.project_id,
    sp.name,
    COALESCE(CAST(GROUP_CONCAT(spp.cidr) AS CHAR), '') as prefixes
FROM
    subnetpools sp
    LEFT JOIN subnetpoolprefixes spp ON sp.id = spp.subnetpool_id
GROUP BY
    sp.id,
    sp.ip_version,
    sp.max_prefixlen,
    sp.min_prefixlen,
    sp.default_prefixlen;

-- name: GetQuotas :many
SELECT
    q.project_id,
    q.resource,
    q.`limit`
FROM
    quotas q
WHERE
    q.project_id IS NOT NULL;

-- name: GetResourceCountsByProject :many
SELECT
    project_id,
    'floatingip' as resource,
    CAST(COUNT(*) AS SIGNED) as cnt
FROM floatingips WHERE project_id IS NOT NULL GROUP BY project_id
UNION ALL
SELECT
    project_id,
    'network' as resource,
    CAST(COUNT(*) AS SIGNED) as cnt
FROM networks WHERE project_id IS NOT NULL GROUP BY project_id
UNION ALL
SELECT
    project_id,
    'port' as resource,
    CAST(COUNT(*) AS SIGNED) as cnt
FROM ports WHERE project_id IS NOT NULL GROUP BY project_id
UNION ALL
SELECT
    project_id,
    'router' as resource,
    CAST(COUNT(*) AS SIGNED) as cnt
FROM routers WHERE project_id IS NOT NULL GROUP BY project_id
UNION ALL
SELECT
    project_id,
    'security_group' as resource,
    CAST(COUNT(*) AS SIGNED) as cnt
FROM securitygroups WHERE project_id IS NOT NULL GROUP BY project_id
UNION ALL
SELECT
    project_id,
    'security_group_rule' as resource,
    CAST(COUNT(*) AS SIGNED) as cnt
FROM securitygrouprules WHERE project_id IS NOT NULL GROUP BY project_id
UNION ALL
SELECT
    project_id,
    'subnet' as resource,
    CAST(COUNT(*) AS SIGNED) as cnt
FROM subnets WHERE project_id IS NOT NULL GROUP BY project_id
UNION ALL
SELECT
    project_id,
    'rbac_policy' as resource,
    CAST(COUNT(*) AS SIGNED) as cnt
FROM networkrbacs WHERE project_id IS NOT NULL GROUP BY project_id
UNION ALL
SELECT
    project_id,
    'subnetpool' as resource,
    CAST(COUNT(*) AS SIGNED) as cnt
FROM subnetpools WHERE project_id IS NOT NULL GROUP BY project_id;
