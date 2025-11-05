CREATE TABLE
    `project` (
        `id` varchar(64) NOT NULL,
        `name` varchar(64) NOT NULL,
        `extra` text,
        `description` text,
        `enabled` tinyint(1) DEFAULT NULL,
        `domain_id` varchar(64) NOT NULL,
        `parent_id` varchar(64) DEFAULT NULL,
        `is_domain` tinyint(1) NOT NULL DEFAULT '0',
        PRIMARY KEY (`id`),
        UNIQUE KEY `ixu_project_name_domain_id` (`domain_id`,`name`),
        KEY `project_parent_id_fkey` (`parent_id`),
        CONSTRAINT `project_domain_id_fkey` FOREIGN KEY (`domain_id`) REFERENCES `project` (`id`),
        CONSTRAINT `project_parent_id_fkey` FOREIGN KEY (`parent_id`) REFERENCES `project` (`id`)
    );

CREATE TABLE
    `region` (
        `id` varchar(255) NOT NULL,
        `description` varchar(255) NOT NULL,
        `parent_region_id` varchar(255) DEFAULT NULL,
        `extra` text,
        PRIMARY KEY (`id`)
    );

CREATE TABLE
    `project_tag` (
        `project_id` varchar(64) NOT NULL,
        `name` varchar(255) NOT NULL,
        PRIMARY KEY (`project_id`,`name`),
        CONSTRAINT `project_tag_ibfk_1` FOREIGN KEY (`project_id`) REFERENCES `project` (`id`) ON DELETE CASCADE
    );

CREATE TABLE
    `user` (
        `id` varchar(64) NOT NULL,
        `extra` text,
        `enabled` tinyint(1) DEFAULT NULL,
        `default_project_id` varchar(64) DEFAULT NULL,
        `created_at` datetime DEFAULT NULL,
        `last_active_at` date DEFAULT NULL,
        `domain_id` varchar(64) NOT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `ixu_user_id_domain_id` (`id`,`domain_id`),
        KEY `ix_default_project_id` (`default_project_id`),
        KEY `domain_id` (`domain_id`)
    );

CREATE TABLE
    `group` (
        `id` varchar(64) NOT NULL,
        `domain_id` varchar(64) NOT NULL,
        `name` varchar(64) NOT NULL,
        `description` text,
        `extra` text,
        PRIMARY KEY (`id`),
        UNIQUE KEY `ixu_group_name_domain_id` (`domain_id`,`name`)
    );
