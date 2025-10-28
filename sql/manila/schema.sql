CREATE TABLE
    `shares` (
        `created_at` datetime(6) DEFAULT NULL,
        `updated_at` datetime(6) DEFAULT NULL,
        `deleted_at` datetime(6) DEFAULT NULL,
        `deleted` varchar(36) DEFAULT NULL,
        `id` varchar(36) NOT NULL,
        `user_id` varchar(255) DEFAULT NULL,
        `project_id` varchar(255) DEFAULT NULL,
        `size` int DEFAULT NULL,
        `display_name` varchar(255) DEFAULT NULL,
        `display_description` varchar(255) DEFAULT NULL,
        `snapshot_id` varchar(36) DEFAULT NULL,
        `share_proto` varchar(255) DEFAULT NULL,
        `is_public` tinyint(1) DEFAULT NULL,
        `snapshot_support` tinyint(1) DEFAULT NULL,
        `share_group_id` varchar(36) DEFAULT NULL,
        `source_share_group_snapshot_member_id` varchar(36) DEFAULT NULL,
        `task_state` varchar(255) DEFAULT NULL,
        `replication_type` varchar(255) DEFAULT NULL,
        `create_share_from_snapshot_support` tinyint(1) DEFAULT NULL,
        `revert_to_snapshot_support` tinyint(1) DEFAULT NULL,
        `mount_snapshot_support` tinyint(1) DEFAULT NULL,
        `is_soft_deleted` tinyint(1) NOT NULL DEFAULT '0',
        `scheduled_to_be_deleted_at` datetime DEFAULT NULL,
        `source_backup_id` varchar(36) DEFAULT NULL,
        PRIMARY KEY (`id`),
        KEY `fk_shares_share_group_id` (`share_group_id`),
        CONSTRAINT `fk_shares_share_group_id` FOREIGN KEY (`share_group_id`) REFERENCES `share_groups` (`id`)
    );

CREATE TABLE
    `share_types` (
        `created_at` datetime(6) DEFAULT NULL,
        `updated_at` datetime(6) DEFAULT NULL,
        `deleted_at` datetime(6) DEFAULT NULL,
        `deleted` varchar(36) DEFAULT NULL,
        `id` varchar(36) NOT NULL,
        `name` varchar(255) DEFAULT NULL,
        `is_public` tinyint(1) DEFAULT NULL,
        `description` varchar(255) DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `st_name_uc` (`name`,`deleted`)
    );

CREATE TABLE
    `availability_zones` (
        `created_at` datetime(6) DEFAULT NULL,
        `updated_at` datetime(6) DEFAULT NULL,
        `deleted_at` datetime(6) DEFAULT NULL,
        `deleted` varchar(36) DEFAULT NULL,
        `id` varchar(36) NOT NULL,
        `name` varchar(255) DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `az_name_uc` (`name`,`deleted`)
    );

CREATE TABLE
    `share_instances` (
        `created_at` datetime(6) DEFAULT NULL,
        `updated_at` datetime(6) DEFAULT NULL,
        `deleted_at` datetime(6) DEFAULT NULL,
        `deleted` varchar(36) DEFAULT NULL,
        `id` varchar(36) NOT NULL,
        `share_id` varchar(36) DEFAULT NULL,
        `host` varchar(255) DEFAULT NULL,
        `status` varchar(255) DEFAULT NULL,
        `scheduled_at` datetime DEFAULT NULL,
        `launched_at` datetime DEFAULT NULL,
        `terminated_at` datetime DEFAULT NULL,
        `share_network_id` varchar(36) DEFAULT NULL,
        `share_server_id` varchar(36) DEFAULT NULL,
        `availability_zone_id` varchar(36) DEFAULT NULL,
        `access_rules_status` varchar(255) DEFAULT NULL,
        `replica_state` varchar(255) DEFAULT NULL,
        `share_type_id` varchar(36) DEFAULT NULL,
        `cast_rules_to_readonly` tinyint(1) NOT NULL,
        `progress` varchar(32) DEFAULT NULL,
        `mount_point_name` varchar(255) DEFAULT NULL,
        PRIMARY KEY (`id`),
        KEY `si_share_network_fk` (`share_network_id`),
        KEY `si_share_server_fk` (`share_server_id`),
        KEY `si_az_id_fk` (`availability_zone_id`),
        KEY `si_st_id_fk` (`share_type_id`),
        KEY `share_instances_share_id_idx` (`share_id`),
        CONSTRAINT `si_az_id_fk` FOREIGN KEY (`availability_zone_id`) REFERENCES `availability_zones` (`id`),
        CONSTRAINT `si_share_fk` FOREIGN KEY (`share_id`) REFERENCES `shares` (`id`),
        CONSTRAINT `si_st_id_fk` FOREIGN KEY (`share_type_id`) REFERENCES `share_types` (`id`)
    );
