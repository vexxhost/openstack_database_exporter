CREATE TABLE
    `agents` (
        `id` varchar(36) NOT NULL PRIMARY KEY,
        `host` varchar(255) NOT NULL,
        `admin_state_up` tinyint (1) NOT NULL DEFAULT '1',
        `heartbeat_timestamp` datetime NOT NULL
    );

CREATE TABLE
    `ha_router_agent_port_bindings` (
        `port_id` varchar(36) NOT NULL PRIMARY KEY,
        `router_id` varchar(36) NOT NULL,
        `l3_agent_id` varchar(36) DEFAULT NULL,
        `state` enum ('active', 'standby', 'unknown') DEFAULT 'standby'
    );
