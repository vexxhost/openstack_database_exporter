CREATE TABLE
    `amphora` (
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `compute_id` varchar(36) DEFAULT NULL,
        `status` varchar(36) NOT NULL,
        `load_balancer_id` varchar(36) DEFAULT NULL,
        `lb_network_ip` varchar(64) DEFAULT NULL,
        `vrrp_ip` varchar(64) DEFAULT NULL,
        `ha_ip` varchar(64) DEFAULT NULL,
        `vrrp_port_id` varchar(36) DEFAULT NULL,
        `ha_port_id` varchar(36) DEFAULT NULL,
        `role` varchar(36) DEFAULT NULL,
        `cert_expiration` datetime DEFAULT NULL,
        `cert_busy` tinyint (1) NOT NULL,
        `vrrp_interface` varchar(16) DEFAULT NULL,
        `vrrp_id` int DEFAULT NULL,
        `vrrp_priority` int DEFAULT NULL,
        `cached_zone` varchar(255) DEFAULT NULL,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `image_id` varchar(36) DEFAULT NULL,
        `compute_flavor` varchar(255) DEFAULT NULL
    );

CREATE TABLE
    `load_balancer` (
        `project_id` varchar(36) DEFAULT NULL,
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `name` varchar(255) DEFAULT NULL,
        `description` varchar(255) DEFAULT NULL,
        `provisioning_status` varchar(16) NOT NULL,
        `operating_status` varchar(16) NOT NULL,
        `enabled` tinyint (1) NOT NULL,
        `topology` varchar(36) DEFAULT NULL,
        `server_group_id` varchar(36) DEFAULT NULL,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `provider` varchar(64) DEFAULT NULL,
        `flavor_id` varchar(36) DEFAULT NULL,
        `availability_zone` varchar(255) DEFAULT NULL
    );

CREATE TABLE
    `pool` (
        `project_id` varchar(36) DEFAULT NULL,
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `name` varchar(255) DEFAULT NULL,
        `description` varchar(255) DEFAULT NULL,
        `protocol` varchar(16) NOT NULL,
        `lb_algorithm` varchar(255) NOT NULL,
        `operating_status` varchar(16) NOT NULL,
        `enabled` tinyint (1) NOT NULL,
        `load_balancer_id` varchar(36) DEFAULT NULL,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `provisioning_status` varchar(16) NOT NULL,
        `tls_certificate_id` varchar(255) DEFAULT NULL,
        `ca_tls_certificate_id` varchar(255) DEFAULT NULL,
        `crl_container_id` varchar(255) DEFAULT NULL,
        `tls_enabled` tinyint (1) NOT NULL DEFAULT '0',
        `tls_ciphers` varchar(2048) DEFAULT NULL,
        `tls_versions` varchar(512) DEFAULT NULL,
        `alpn_protocols` varchar(512) DEFAULT NULL
    );

CREATE TABLE
    `vip` (
        `load_balancer_id` varchar(36) NOT NULL PRIMARY KEY,
        `ip_address` varchar(64) DEFAULT NULL,
        `port_id` varchar(36) DEFAULT NULL,
        `subnet_id` varchar(36) DEFAULT NULL,
        `network_id` varchar(36) DEFAULT NULL,
        `qos_policy_id` varchar(36) DEFAULT NULL,
        `octavia_owned` tinyint (1) DEFAULT NULL,
        `vnic_type` varchar(64) NOT NULL DEFAULT 'normal'
    );
