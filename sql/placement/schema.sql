CREATE TABLE
    `resource_providers` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `uuid` varchar(36) NOT NULL,
        `name` varchar(200) DEFAULT NULL,
        `generation` int(11) NOT NULL,
        `can_host` int(11) NOT NULL DEFAULT '0',
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        `root_provider_id` int(11) NOT NULL,
        `parent_provider_id` int(11) DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `uniq_resource_providers0uuid` (`uuid`),
        KEY `resource_providers_name_idx` (`name`),
        KEY `resource_providers_root_provider_id_idx` (`root_provider_id`),
        KEY `resource_providers_parent_provider_id_idx` (`parent_provider_id`)
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
        KEY `allocations_consumer_id_idx` (`consumer_id`),
        CONSTRAINT `allocations_ibfk_1` FOREIGN KEY (`resource_provider_id`) REFERENCES `resource_providers` (`id`),
        CONSTRAINT `allocations_ibfk_2` FOREIGN KEY (`resource_class_id`) REFERENCES `resource_classes` (`id`)
    );

CREATE TABLE
    `inventories` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `resource_provider_id` int(11) NOT NULL,
        `resource_class_id` int(11) NOT NULL,
        `total` int(11) NOT NULL,
        `reserved` int(11) NOT NULL DEFAULT '0',
        `min_unit` int(11) NOT NULL DEFAULT '1',
        `max_unit` int(11) NOT NULL,
        `step_size` int(11) NOT NULL DEFAULT '1',
        `allocation_ratio` decimal(16,4) NOT NULL DEFAULT '1.0000',
        `created_at` datetime DEFAULT NULL,
        `updated_at` datetime DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `uniq_inventories0resource_provider_resource_class` (`resource_provider_id`,`resource_class_id`),
        KEY `inventories_resource_class_id_idx` (`resource_class_id`),
        CONSTRAINT `inventories_ibfk_1` FOREIGN KEY (`resource_provider_id`) REFERENCES `resource_providers` (`id`),
        CONSTRAINT `inventories_ibfk_2` FOREIGN KEY (`resource_class_id`) REFERENCES `resource_classes` (`id`)
    );
