CREATE TABLE
    `quotas` (
        `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `deleted_at` datetime DEFAULT NULL,
        `deleted` tinyint (1) DEFAULT NULL,
        `project_id` varchar(255) DEFAULT NULL,
        `resource` varchar(300) NOT NULL,
        `hard_limit` int DEFAULT NULL,
        `allocated` int DEFAULT NULL
    );

CREATE TABLE
    `quota_usages` (
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `deleted_at` datetime DEFAULT NULL,
        `deleted` tinyint (1) DEFAULT NULL,
        `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
        `project_id` varchar(255) DEFAULT NULL,
        `resource` varchar(300) DEFAULT NULL,
        `in_use` int NOT NULL,
        `reserved` int NOT NULL,
        `until_refresh` int DEFAULT NULL,
        `race_preventer` tinyint (1) DEFAULT NULL
    );

CREATE TABLE
    `services` (
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `deleted_at` datetime DEFAULT NULL,
        `deleted` tinyint (1) DEFAULT NULL,
        `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
        `host` varchar(255) DEFAULT NULL,
        `binary` varchar(255) DEFAULT NULL,
        `topic` varchar(255) DEFAULT NULL,
        `report_count` int NOT NULL,
        `disabled` tinyint (1) DEFAULT NULL,
        `availability_zone` varchar(255) DEFAULT NULL,
        `disabled_reason` varchar(255) DEFAULT NULL,
        `modified_at` datetime DEFAULT NULL,
        `rpc_current_version` varchar(36) DEFAULT NULL,
        `object_current_version` varchar(36) DEFAULT NULL,
        `replication_status` varchar(36) DEFAULT NULL,
        `frozen` tinyint (1) DEFAULT NULL,
        `active_backend_id` varchar(255) DEFAULT NULL,
        `cluster_name` varchar(255) DEFAULT NULL,
        `uuid` varchar(36) DEFAULT NULL
    );

CREATE TABLE
    `snapshots` (
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `deleted_at` datetime DEFAULT NULL,
        `deleted` tinyint (1) DEFAULT NULL,
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `volume_id` varchar(36) NOT NULL,
        `user_id` varchar(255) DEFAULT NULL,
        `project_id` varchar(255) DEFAULT NULL,
        `status` varchar(255) DEFAULT NULL,
        `progress` varchar(255) DEFAULT NULL,
        `volume_size` int DEFAULT NULL,
        `scheduled_at` datetime DEFAULT NULL,
        `display_name` varchar(255) DEFAULT NULL,
        `display_description` varchar(255) DEFAULT NULL,
        `provider_location` varchar(255) DEFAULT NULL,
        `encryption_key_id` varchar(36) DEFAULT NULL,
        `volume_type_id` varchar(36) NOT NULL,
        `cgsnapshot_id` varchar(36) DEFAULT NULL,
        `provider_id` varchar(255) DEFAULT NULL,
        `provider_auth` varchar(255) DEFAULT NULL,
        `group_snapshot_id` varchar(36) DEFAULT NULL,
        `use_quota` tinyint (1) NOT NULL DEFAULT '1'
    );

CREATE TABLE
    `volumes` (
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `deleted_at` datetime DEFAULT NULL,
        `deleted` tinyint (1) DEFAULT NULL,
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `ec2_id` varchar(255) DEFAULT NULL,
        `user_id` varchar(255) DEFAULT NULL,
        `project_id` varchar(255) DEFAULT NULL,
        `host` varchar(255) DEFAULT NULL,
        `size` int DEFAULT NULL,
        `availability_zone` varchar(255) DEFAULT NULL,
        `status` varchar(255) DEFAULT NULL,
        `attach_status` varchar(255) DEFAULT NULL,
        `scheduled_at` datetime DEFAULT NULL,
        `launched_at` datetime DEFAULT NULL,
        `terminated_at` datetime DEFAULT NULL,
        `display_name` varchar(255) DEFAULT NULL,
        `display_description` varchar(255) DEFAULT NULL,
        `provider_location` varchar(256) DEFAULT NULL,
        `provider_auth` varchar(256) DEFAULT NULL,
        `snapshot_id` varchar(36) DEFAULT NULL,
        `volume_type_id` varchar(36) NOT NULL,
        `source_volid` varchar(36) DEFAULT NULL,
        `bootable` tinyint (1) DEFAULT NULL,
        `provider_geometry` varchar(255) DEFAULT NULL,
        `_name_id` varchar(36) DEFAULT NULL,
        `encryption_key_id` varchar(36) DEFAULT NULL,
        `migration_status` varchar(255) DEFAULT NULL,
        `replication_status` varchar(255) DEFAULT NULL,
        `replication_extended_status` varchar(255) DEFAULT NULL,
        `replication_driver_data` varchar(255) DEFAULT NULL,
        `consistencygroup_id` varchar(36) DEFAULT NULL,
        `provider_id` varchar(255) DEFAULT NULL,
        `multiattach` tinyint (1) DEFAULT NULL,
        `previous_status` varchar(255) DEFAULT NULL,
        `cluster_name` varchar(255) DEFAULT NULL,
        `group_id` varchar(36) DEFAULT NULL,
        `service_uuid` varchar(36) DEFAULT NULL,
        `shared_targets` tinyint (1) DEFAULT NULL,
        `use_quota` tinyint (1) NOT NULL DEFAULT '1'
    );

CREATE TABLE
    `volume_attachment` (
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `deleted_at` datetime DEFAULT NULL,
        `deleted` tinyint (1) DEFAULT NULL,
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `volume_id` varchar(36) NOT NULL,
        `attached_host` varchar(255) DEFAULT NULL,
        `instance_uuid` varchar(36) DEFAULT NULL,
        `mountpoint` varchar(255) DEFAULT NULL,
        `attach_time` datetime DEFAULT NULL,
        `detach_time` datetime DEFAULT NULL,
        `attach_mode` varchar(36) DEFAULT NULL,
        `attach_status` varchar(255) DEFAULT NULL,
        `connection_info` text,
        `connector` text
    );

CREATE TABLE
    `volume_types` (
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `deleted_at` datetime DEFAULT NULL,
        `deleted` tinyint (1) DEFAULT NULL,
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `name` varchar(255) DEFAULT NULL,
        `qos_specs_id` varchar(36) DEFAULT NULL,
        `is_public` tinyint (1) DEFAULT NULL,
        `description` varchar(255) DEFAULT NULL
    );
