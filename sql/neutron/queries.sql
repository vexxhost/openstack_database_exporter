-- name: GetHARouterAgentPortBindingsWithAgents :many
SELECT
    ha.router_id,
    ha.l3_agent_id,
    ha.state,
    a.host as agent_host,
    a.admin_state_up as agent_admin_state_up,
    a.heartbeat_timestamp as agent_heartbeat_timestamp
FROM
    ha_router_agent_port_bindings ha
    LEFT JOIN agents a ON ha.l3_agent_id = a.id;
