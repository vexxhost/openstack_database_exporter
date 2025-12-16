-- name: GetNodeMetrics :many
SELECT 
    uuid,
    name,
    power_state,
    provision_state,
    maintenance,
    resource_class,
    console_enabled,
    retired,
    COALESCE(retired_reason, '') as retired_reason
FROM nodes
WHERE provision_state != 'deleted'
ORDER BY created_at;
