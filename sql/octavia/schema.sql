CREATE TABLE
    `amphora` (
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `compute_id` varchar(36) DEFAULT NULL,
        `status` varchar(36) NOT NULL,
        `load_balancer_id` varchar(36) DEFAULT NULL,
        `lb_network_ip` varchar(64) DEFAULT NULL,
        `ha_ip` varchar(64) DEFAULT NULL,
        `role` varchar(36) DEFAULT NULL,
        `cert_expiration` datetime DEFAULT NULL
    );

CREATE TABLE
    `load_balancer` (
        `project_id` varchar(36) DEFAULT NULL,
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `name` varchar(255) DEFAULT NULL,
        `provisioning_status` varchar(16) NOT NULL,
        `operating_status` varchar(16) NOT NULL,
        `provider` varchar(64) DEFAULT NULL
    );

CREATE TABLE
    `pool` (
        `project_id` varchar(36) DEFAULT NULL,
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `name` varchar(255) DEFAULT NULL,
        `protocol` varchar(16) NOT NULL,
        `lb_algorithm` varchar(255) NOT NULL,
        `operating_status` varchar(16) NOT NULL,
        `load_balancer_id` varchar(36) DEFAULT NULL,
        `provisioning_status` varchar(16) NOT NULL
    );

CREATE TABLE
    `vip` (
        `load_balancer_id` varchar(36) NOT NULL PRIMARY KEY,
        `ip_address` varchar(64) DEFAULT NULL
    );
