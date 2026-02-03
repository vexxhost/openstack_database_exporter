CREATE TABLE
    `agents` (
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `host` varchar(255) NOT NULL,
        `admin_state_up` tinyint (1) NOT NULL DEFAULT '1',
        `heartbeat_timestamp` datetime NOT NULL
    );

CREATE TABLE
    `ha_router_agent_port_bindings` (
        `port_id` varchar(36) NOT NULL PRIMARY KEY,
        `router_id` varchar(36) NOT NULL,
        `l3_agent_id` varchar(36) DEFAULT NULL,
        `state` enum ('active', 'standby', 'unknown') DEFAULT 'standby'
    );

CREATE TABLE
    `routers` (
        `project_id` varchar(255) DEFAULT NULL,
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `name` varchar(255) DEFAULT NULL,
        `status` varchar(16) DEFAULT NULL,
        `admin_state_up` tinyint(1) DEFAULT NULL,
        `gw_port_id` varchar(36) DEFAULT NULL,
        `enable_snat` tinyint(1) NOT NULL DEFAULT '1',
        `standard_attr_id` bigint NOT NULL,
        `flavor_id` varchar(36) DEFAULT NULL
    );

CREATE TABLE
    `floatingips` (
        `project_id` varchar(255) DEFAULT NULL,
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `floating_ip_address` varchar(64) NOT NULL,
        `floating_network_id` varchar(36) NOT NULL,
        `floating_port_id` varchar(36) NOT NULL,
        `fixed_port_id` varchar(36) DEFAULT NULL,
        `fixed_ip_address` varchar(64) DEFAULT NULL,
        `router_id` varchar(36) DEFAULT NULL,
        `last_known_router_id` varchar(36) DEFAULT NULL,
        `status` varchar(16) DEFAULT NULL,
        `standard_attr_id` bigint NOT NULL
    );

CREATE TABLE
    `networks` (
        `project_id` varchar(255) DEFAULT NULL,
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `name` varchar(255) DEFAULT NULL,
        `status` varchar(16) DEFAULT NULL,
        `admin_state_up` tinyint(1) DEFAULT NULL,
        `vlan_transparent` tinyint(1) DEFAULT NULL,
        `standard_attr_id` bigint NOT NULL,
        `availability_zone_hints` varchar(255) DEFAULT NULL,
        `mtu` int NOT NULL DEFAULT '1500'
    );

CREATE TABLE
    `networksegments` (
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `network_id` varchar(36) NOT NULL,
        `network_type` varchar(32) NOT NULL,
        `physical_network` varchar(64) DEFAULT NULL,
        `segmentation_id` int DEFAULT NULL,
        `is_dynamic` tinyint(1) NOT NULL DEFAULT '0',
        `segment_index` int NOT NULL DEFAULT '0',
        `standard_attr_id` bigint NOT NULL,
        `name` varchar(255) DEFAULT NULL
    );

CREATE TABLE
    `subnets` (
        `project_id` varchar(255) DEFAULT NULL,
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `name` varchar(255) DEFAULT NULL,
        `network_id` varchar(36) NOT NULL,
        `ip_version` int NOT NULL,
        `cidr` varchar(64) NOT NULL,
        `gateway_ip` varchar(64) DEFAULT NULL,
        `enable_dhcp` tinyint(1) DEFAULT NULL,
        `ipv6_ra_mode` enum('slaac','dhcpv6-stateful','dhcpv6-stateless') DEFAULT NULL,
        `ipv6_address_mode` enum('slaac','dhcpv6-stateful','dhcpv6-stateless') DEFAULT NULL,
        `subnetpool_id` varchar(36) DEFAULT NULL,
        `standard_attr_id` bigint NOT NULL,
        `segment_id` varchar(36) DEFAULT NULL
    );

CREATE TABLE 
    `externalnetworks` (
        `network_id` varchar(36) NOT NULL PRIMARY KEY,
        `is_default` tinyint(1) NOT NULL DEFAULT '0'
);

CREATE TABLE
    `ml2_port_bindings` (
        `port_id` varchar(36) NOT NULL,
        `host` varchar(255) NOT NULL DEFAULT '',
        `vif_type` varchar(64) NOT NULL,
        `vnic_type` varchar(64) NOT NULL DEFAULT 'normal',
        `profile` varchar(4095) NOT NULL DEFAULT '',
        `vif_details` varchar(4095) NOT NULL DEFAULT '',
        `status` varchar(16) NOT NULL DEFAULT 'ACTIVE',
        PRIMARY KEY (`port_id`, `host`)
    );

CREATE TABLE 
    `ports` (
        `project_id` varchar(255) DEFAULT NULL,
        `id` varchar(36) NOT NULL,
        `name` varchar(255) DEFAULT NULL,
        `network_id` varchar(36) NOT NULL,
        `mac_address` varchar(32) NOT NULL,
        `admin_state_up` tinyint(1) NOT NULL,
        `status` varchar(16) NOT NULL,
        `device_id` varchar(255) NOT NULL,
        `device_owner` varchar(255) NOT NULL,
        `standard_attr_id` bigint NOT NULL,
        `ip_allocation` varchar(16) DEFAULT NULL
);

CREATE TABLE 
    `securitygroups` (
        `project_id` varchar(255) DEFAULT NULL,
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `name` varchar(255) DEFAULT NULL,
        `standard_attr_id` bigint NOT NULL,
        `stateful` tinyint(1) NOT NULL DEFAULT '1'
);

CREATE TABLE 
    `dnsnameservers` (
        `address` varchar(128) NOT NULL,
        `subnet_id` varchar(36) NOT NULL,
        `order` int NOT NULL DEFAULT '0',
        PRIMARY KEY (`address`,`subnet_id`),
        KEY `subnet_id` (`subnet_id`)
);

CREATE TABLE 
    `ipallocations` (
        `port_id` varchar(36) DEFAULT NULL,
        `ip_address` varchar(64) NOT NULL,
        `subnet_id` varchar(36) NOT NULL,
        `network_id` varchar(36) NOT NULL,
        PRIMARY KEY (`ip_address`,`subnet_id`,`network_id`)
);

CREATE TABLE `networkrbacs` (
    `id` varchar(36) NOT NULL,
        `object_id` varchar(36) NOT NULL,
        `project_id` varchar(255) DEFAULT NULL,
        `target_project` varchar(255) NOT NULL,
        `action` varchar(255) NOT NULL,
PRIMARY KEY (`id`)
);

CREATE TABLE 
    `subnetpools` (
        `project_id` varchar(255) DEFAULT NULL,
        `id` varchar(36) NOT NULL,
        `name` varchar(255) DEFAULT NULL,
        `ip_version` int NOT NULL,
        `default_prefixlen` int NOT NULL,
        `min_prefixlen` int NOT NULL,
        `max_prefixlen` int NOT NULL,
        `shared` tinyint(1) NOT NULL DEFAULT '0',
        `default_quota` int DEFAULT NULL,
        `hash` varchar(36) NOT NULL DEFAULT '',
        `address_scope_id` varchar(36) DEFAULT NULL,
        `is_default` tinyint(1) NOT NULL DEFAULT '0',
        `standard_attr_id` bigint NOT NULL,
PRIMARY KEY (`id`)
    );

CREATE TABLE 
    `subnetpoolprefixes` (
        `cidr` varchar(64) NOT NULL,
        `subnetpool_id` varchar(36) NOT NULL,
        PRIMARY KEY (`cidr`,`subnetpool_id`),
        KEY `subnetpool_id` (`subnetpool_id`)
);

CREATE TABLE 
    `ipallocationpools` (
        `id` varchar(36) NOT NULL,
        `subnet_id` varchar(36) DEFAULT NULL,
        `first_ip` varchar(64) NOT NULL,
        `last_ip` varchar(64) NOT NULL,
        PRIMARY KEY (`id`)
);