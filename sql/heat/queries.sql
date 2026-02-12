-- name: GetStackMetrics :many
SELECT 
    s.id,
    COALESCE(s.name, '') as name,
    COALESCE(s.status, '') as status,
    COALESCE(s.action, '') as action,
    COALESCE(s.tenant, '') as tenant
FROM stack s
WHERE s.deleted_at IS NULL;
