-- name: GetAllServices :many
SELECT
    uuid,
    host,
    `binary` as service,
    CASE
        WHEN disabled = 1 THEN 'disabled'
        ELSE 'enabled'
    END as admin_state,
    availability_zone as zone,
    disabled_reason,
    CASE
        WHEN TIMESTAMPDIFF (SECOND, updated_at, NOW()) <= 60 THEN 1
        ELSE 0
    END as state
FROM
    services
WHERE
    deleted = 0;

-- name: GetProjectQuotaLimits :many
SELECT
    q.project_id,
    q.resource,
    q.hard_limit,
    COALESCE(qu.in_use, 0) as in_use
FROM
    quotas q
    LEFT JOIN quota_usages qu ON q.project_id = qu.project_id
        AND q.resource = qu.resource
        AND qu.deleted = 0
WHERE
    q.deleted = 0
    AND q.resource IN ('gigabytes', 'backup_gigabytes');

-- name: GetSnapshotCount :one
SELECT
    COUNT(*) as count
FROM
    snapshots
WHERE
    deleted = 0;

-- name: GetAllVolumes :many
SELECT
    v.id,
    v.display_name as name,
    v.size,
    v.status,
    v.availability_zone,
    v.bootable,
    v.project_id,
    v.user_id,
    vt.name as volume_type,
    va.instance_uuid as server_id
FROM
    volumes v USE INDEX (volumes_service_uuid_idx)
    LEFT JOIN volume_types vt ON v.volume_type_id = vt.id
    LEFT JOIN volume_attachment va ON v.id = va.volume_id AND va.deleted = 0
WHERE
    (v.service_uuid IS NULL OR v.service_uuid IS NOT NULL)
    AND v.deleted = 0;
