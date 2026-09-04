CREATE TABLE
    `failover_segments` (
        `id` int NOT NULL AUTO_INCREMENT,
        `uuid` varchar(36) NOT NULL,
        `name` varchar(255) NOT NULL,
        `service_type` varchar(255) NOT NULL,
        `enabled` tinyint(1) NOT NULL DEFAULT '1',
        `description` text,
        `recovery_method` enum('auto','reserved_host','auto_priority','rh_priority') NOT NULL,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `deleted_at` datetime DEFAULT NULL,
        `deleted` tinyint(1) NOT NULL DEFAULT '0',
        PRIMARY KEY (`id`),
        UNIQUE KEY `uniq_segment0name0deleted` (`name`, `deleted`),
        UNIQUE KEY `uniq_segments0uuid` (`uuid`),
        KEY `segments_service_type_idx` (`service_type`)
    );

CREATE TABLE
    `hosts` (
        `id` int NOT NULL AUTO_INCREMENT,
        `uuid` varchar(36) NOT NULL,
        `name` varchar(255) NOT NULL,
        `reserved` tinyint(1) NOT NULL DEFAULT '0',
        `type` varchar(255) NOT NULL,
        `control_attributes` text NOT NULL,
        `on_maintenance` tinyint(1) NOT NULL DEFAULT '0',
        `failover_segment_id` varchar(36) NOT NULL,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `deleted_at` datetime DEFAULT NULL,
        `deleted` tinyint(1) NOT NULL DEFAULT '0',
        PRIMARY KEY (`id`),
        UNIQUE KEY `uniq_host0name0deleted` (`name`, `deleted`),
        UNIQUE KEY `uniq_host0uuid` (`uuid`),
        KEY `hosts_type_idx` (`type`)
    );

CREATE TABLE
    `notifications` (
        `id` int NOT NULL AUTO_INCREMENT,
        `notification_uuid` varchar(36) NOT NULL,
        `generated_time` datetime NOT NULL,
        `type` varchar(36) NOT NULL,
        `payload` text,
        `status` enum('new','running','error','failed','ignored','finished') NOT NULL,
        `source_host_uuid` varchar(36) NOT NULL,
        `failover_segment_uuid` varchar(36) NOT NULL,
        `message` text,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `deleted_at` datetime DEFAULT NULL,
        `deleted` tinyint(1) NOT NULL DEFAULT '0',
        PRIMARY KEY (`id`),
        UNIQUE KEY `uniq_notification0uuid` (`notification_uuid`)
    );
