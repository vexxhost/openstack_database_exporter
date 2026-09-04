-- name: GetZones :many
SELECT
    id,
    name,
    tenant_id,
    type,
    status
FROM
    zones
WHERE
    deleted = '0';

-- name: GetRecordsets :many
SELECT DISTINCT
    r.id,
    r.name,
    r.type,
    r.zone_id,
    z.name AS zone_name,
    r.tenant_id,
    COALESCE(
        (SELECT rec.status FROM records rec
         WHERE rec.recordset_id = r.id
         ORDER BY FIELD(rec.status, 'ERROR', 'PENDING', 'DELETED', 'ACTIVE') DESC
         LIMIT 1),
        'PENDING'
    ) AS status
FROM
    recordsets r
    INNER JOIN zones z ON r.zone_id = z.id AND z.deleted = '0';
