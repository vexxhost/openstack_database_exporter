-- name: GetStackMetrics :many
SELECT 
    s.id,
    COALESCE(s.name, '') as name,
    COALESCE(s.status, '') as status,
    COALESCE(s.action, '') as action,
    COALESCE(s.tenant, '') as tenant,
    s.created_at,
    s.updated_at,
    s.deleted_at,
    COALESCE(s.nested_depth, 0) as nested_depth,
    COALESCE(s.disable_rollback, false) as disable_rollback
FROM stack s
WHERE s.deleted_at IS NULL
ORDER BY s.created_at DESC;
