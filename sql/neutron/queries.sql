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
    r.gw_port_id
FROM
    routers r;

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
    ns.segmentation_id as provider_segmentation_id,
    CAST(GROUP_CONCAT(s.id) as CHAR) as subnets,
    CASE
        WHEN en.network_id IS NOT NULL THEN TRUE
        ELSE FALSE
    END AS is_external,
    CASE
        WHEN rbacs.object_id IS NOT NULL THEN TRUE
        ELSE FALSE
    END AS is_shared
FROM
    networks n
    LEFT JOIN networksegments ns ON n.id = ns.network_id
    LEFT JOIN subnets s on n.id = s.network_id
    LEFT JOIN externalnetworks en on n.id = en.network_id
    LEFT JOIN networkrbacs rbacs on n.id = rbacs.object_id
GROUP BY
    n.id,
    n.name,
    n.project_id,
    n.status,
    ns.network_type,
    ns.physical_network,
    ns.segmentation_id;

-- name: GetSubnets :many
SELECT
    s.id,
    s.cidr,
    s.gateway_ip,
    s.network_id,
    s.project_id,
    s.enable_dhcp,
    CAST(GROUP_CONCAT(d.address) as CHAR) as dns_nameservers,
    s.subnetpool_id
FROM
    subnets s
    LEFT JOIN dnsnameservers d on s.id = d.subnet_id
GROUP BY
    s.id,
    s.cidr,
    s.gateway_ip,
    s.network_id,
    s.project_id,
    s.enable_dhcp;


-- name: GetPorts :many
SELECT
    p.id,
    p.mac_address,
    p.device_owner,
    p.status,
    p.network_id,
    p.admin_state_up,
    b.vif_type as binding_vif_type,
    CAST(GROUP_CONCAT(ia.ip_address) as CHAR) as fixed_ips
FROM
    ports p
    LEFT JOIN ml2_port_bindings b ON p.id = b.port_id
    LEFT JOIN ipallocations ia on p.id = ia.port_id
GROUP BY
    p.id,
    p.mac_address,
    p.device_owner,
    p.status,
    p.network_id,
    p.admin_state_up,
    b.vif_type;

-- name: GetSecurityGroups :many
SELECT
    s.id
FROM
    securitygroups s;

-- name: GetNetworkIPAvailabilitiesUsed :many
SELECT 
	s.id AS subnet_id, 
	s.name AS subnet_name, 
	s.cidr, 
	s.ip_version, 
	s.project_id, 
	n.id AS network_id, 
	n.name AS network_name, 
	COUNT(ipa.ip_address) AS allocation_count
FROM subnets s 
	LEFT JOIN ipallocations ipa ON ipa.subnet_id = s.id 
	LEFT JOIN networks n ON s.network_id = n.id 
GROUP BY s.id, n.id;

-- name: GetSubnetPools :many
SELECT
    sp.id,
    sp.ip_version,
    sp.max_prefixlen,
    sp.min_prefixlen,
    sp.default_prefixlen,
    sp.project_id,
    sp.name,
    CAST(GROUP_CONCAT(spp.cidr) as CHAR) as prefixes
FROM 
    subnetpools sp
    LEFT JOIN subnetpoolprefixes spp on sp.id = spp.subnetpool_id
GROUP BY
    sp.id,
    sp.ip_version,
    sp.max_prefixlen,
    sp.min_prefixlen,
    sp.default_prefixlen;

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

