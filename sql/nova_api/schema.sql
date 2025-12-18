-- Nova API database schema  
-- This schema contains flavors, quotas, and aggregates tables

CREATE TABLE IF NOT EXISTS
    `flavors` (
        `created_at` DATETIME NULL,
        `updated_at` DATETIME NULL,
        `name` VARCHAR(255) NOT NULL,
        `id` INT NOT NULL AUTO_INCREMENT,
        `memory_mb` INT NOT NULL,
        `vcpus` INT NOT NULL,
        `swap` INT NOT NULL,
        `vcpu_weight` INT NULL,
        `flavorid` VARCHAR(255) NOT NULL,
        `rxtx_factor` FLOAT NULL,
        `root_gb` INT NULL,
        `ephemeral_gb` INT NULL,
        `disabled` TINYINT(1) NULL,
        `is_public` TINYINT(1) NULL,
        `description` TEXT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY uniq_flavors0flavorid (`flavorid`),
        UNIQUE KEY uniq_flavors0name (`name`)
    );

CREATE TABLE IF NOT EXISTS
    `quotas` (
        `id` INT NOT NULL AUTO_INCREMENT,
        `created_at` DATETIME NULL,
        `updated_at` DATETIME NULL,
        `project_id` VARCHAR(255) NULL,
        `resource` VARCHAR(255) NOT NULL,
        `hard_limit` INT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY uniq_quotas0project_id0resource (`project_id`, `resource`),
        KEY quotas_project_id_idx (`project_id`)
    );

CREATE TABLE IF NOT EXISTS
    `aggregates` (
        `created_at` DATETIME NULL,
        `updated_at` DATETIME NULL,
        `id` INT NOT NULL AUTO_INCREMENT,
        `uuid` VARCHAR(36) NULL,
        `name` VARCHAR(255) NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY uniq_aggregate0name (`name`),
        KEY aggregate_uuid_idx (`uuid`)
    );

CREATE TABLE IF NOT EXISTS
    `aggregate_hosts` (
        `created_at` DATETIME NULL,
        `updated_at` DATETIME NULL,
        `id` INT NOT NULL AUTO_INCREMENT,
        `host` VARCHAR(255) NULL,
        `aggregate_id` INT NOT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY uniq_aggregate_hosts0host0aggregate_id (`host`, `aggregate_id`),
        KEY aggregate_id (`aggregate_id`),
        CONSTRAINT aggregate_hosts_ibfk_1 FOREIGN KEY (`aggregate_id`) REFERENCES `aggregates` (`id`)
    );

CREATE TABLE IF NOT EXISTS
    `quota_usages` (
        `created_at` DATETIME NULL,
        `updated_at` DATETIME NULL,
        `id` INT NOT NULL AUTO_INCREMENT,
        `project_id` VARCHAR(255) NULL,
        `user_id` VARCHAR(255) NULL,
        `resource` VARCHAR(255) NOT NULL,
        `in_use` INT NOT NULL,
        `reserved` INT NOT NULL,
        `until_refresh` INT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY uniq_quota_usages0project_id0user_id0resource (`project_id`, `user_id`, `resource`),
        KEY quota_usages_project_id_idx (`project_id`),
        KEY quota_usages_user_id_idx (`user_id`)
    );
