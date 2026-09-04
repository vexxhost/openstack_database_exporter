-- name: GetFailoverSegments :many
SELECT
    id,
    uuid,
    name,
    service_type,
    enabled,
    description,
    recovery_method
FROM
    failover_segments
WHERE
    deleted = 0;

-- name: GetHosts :many
SELECT
    h.id,
    h.uuid,
    h.name,
    h.reserved,
    h.type,
    h.control_attributes,
    h.on_maintenance,
    h.failover_segment_id,
    s.name AS failover_segment_name
FROM
    hosts h
    LEFT JOIN failover_segments s ON h.failover_segment_id = s.uuid
        AND s.deleted = 0
WHERE
    h.deleted = 0;

-- name: GetNotificationsByStatus :many
SELECT
    status,
    COUNT(*) AS count
FROM
    notifications
WHERE
    deleted = 0
GROUP BY
    status;
