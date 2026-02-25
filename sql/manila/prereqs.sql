-- Manila schema prerequisites: stub tables referenced by foreign keys
-- but not needed by our queries.
CREATE TABLE IF NOT EXISTS `share_groups` (
    `id` varchar(36) NOT NULL,
    `created_at` datetime(6) DEFAULT NULL,
    `deleted` varchar(36) DEFAULT NULL,
    PRIMARY KEY (`id`)
);
