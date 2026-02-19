CREATE TABLE
    `resource_providers` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `uuid` varchar(36) NOT NULL,
        `name` varchar(200) DEFAULT NULL,
        `generation` int(11) DEFAULT NULL,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `root_provider_id` int(11) DEFAULT NULL,
        `parent_provider_id` int(11) DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `uniq_resource_providers0uuid` (`uuid`),
        UNIQUE KEY `uniq_resource_providers0name` (`name`),
        KEY `resource_providers_root_provider_id_idx` (`root_provider_id`),
        KEY `resource_providers_parent_provider_id_idx` (`parent_provider_id`),
        CONSTRAINT `resource_providers_ibfk_1` FOREIGN KEY (`parent_provider_id`) REFERENCES `resource_providers` (`id`),
        CONSTRAINT `resource_providers_ibfk_2` FOREIGN KEY (`root_provider_id`) REFERENCES `resource_providers` (`id`)
    );

CREATE TABLE
    `resource_classes` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `name` varchar(255) NOT NULL,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `uniq_resource_classes0name` (`name`)
    );

CREATE TABLE
    `allocations` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `resource_provider_id` int(11) NOT NULL,
        `consumer_id` varchar(36) NOT NULL,
        `resource_class_id` int(11) NOT NULL,
        `used` int(11) NOT NULL,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        PRIMARY KEY (`id`),
        KEY `allocations_resource_provider_class_used_idx` (`resource_provider_id`,`resource_class_id`,`used`),
        KEY `allocations_resource_class_id_idx` (`resource_class_id`),
        KEY `allocations_consumer_id_idx` (`consumer_id`)
    );

CREATE TABLE
    `inventories` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `resource_provider_id` int(11) NOT NULL,
        `resource_class_id` int(11) NOT NULL,
        `total` int(11) NOT NULL,
        `reserved` int(11) NOT NULL,
        `min_unit` int(11) NOT NULL,
        `max_unit` int(11) NOT NULL,
        `step_size` int(11) NOT NULL,
        `allocation_ratio` float NOT NULL,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `uniq_inventories0resource_provider_resource_class` (`resource_provider_id`,`resource_class_id`),
        KEY `inventories_resource_class_id_idx` (`resource_class_id`)
    );

CREATE TABLE
    `projects` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `external_id` varchar(255) NOT NULL,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `uniq_projects0external_id` (`external_id`)
    );

CREATE TABLE
    `users` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `external_id` varchar(255) NOT NULL,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `uniq_users0external_id` (`external_id`)
    );

CREATE TABLE
    `consumers` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `uuid` varchar(36) NOT NULL,
        `project_id` int(11) NOT NULL,
        `user_id` int(11) NOT NULL,
        `generation` int(11) NOT NULL DEFAULT '0',
        `consumer_type_id` int(11) DEFAULT NULL,
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `uniq_consumers0uuid` (`uuid`),
        KEY `consumers_project_id_user_id_uuid_idx` (`project_id`,`user_id`,`uuid`),
        KEY `consumers_project_id_uuid_idx` (`project_id`,`uuid`),
        CONSTRAINT `consumers_ibfk_1` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`),
        CONSTRAINT `consumers_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
    );
