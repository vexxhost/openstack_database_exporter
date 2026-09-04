-- name: GetProjectMetrics :many
SELECT 
    p.id,
    p.name,
    COALESCE(p.description, '') as description,
    p.enabled,
    p.domain_id,
    COALESCE(p.parent_id, '') as parent_id,
    p.is_domain,
    CAST(COALESCE(GROUP_CONCAT(pt.name SEPARATOR ','), '') AS CHAR) as tags
FROM project p
LEFT JOIN project_tag pt ON p.id = pt.project_id
WHERE p.is_domain = 0
GROUP BY p.id, p.name, p.description, p.enabled, p.domain_id, p.parent_id, p.is_domain;

-- name: GetDomainMetrics :many
SELECT 
    p.id,
    p.name,
    COALESCE(p.description, '') as description,
    p.enabled,
    CAST(COALESCE(GROUP_CONCAT(pt.name SEPARATOR ','), '') AS CHAR) as tags
FROM project p
LEFT JOIN project_tag pt ON p.id = pt.project_id
WHERE p.is_domain = 1 AND p.id != '<<keystone.domain.root>>'
GROUP BY p.id, p.name, p.description, p.enabled;

-- name: GetUserMetrics :many
SELECT 
    id,
    enabled,
    domain_id,
    COALESCE(default_project_id, '') as default_project_id,
    created_at,
    last_active_at
FROM user;

-- name: GetRegionMetrics :many
SELECT 
    id,
    COALESCE(description, '') as description,
    COALESCE(parent_region_id, '') as parent_region_id
FROM region;

-- name: GetGroupMetrics :many
SELECT 
    id,
    domain_id,
    name,
    COALESCE(description, '') as description
FROM `group`;

-- name: GetRegisteredLimits :many
SELECT DISTINCT
    rl.resource_name,
    rl.default_limit
FROM registered_limit rl
INNER JOIN service s ON rl.service_id = s.id
WHERE s.type = 'compute'
  AND (rl.region_id = sqlc.arg(region_id) OR (rl.region_id IS NULL AND sqlc.arg(region_id) = ''));

-- name: GetProjectLimits :many
SELECT
    l.project_id,
    rl.resource_name,
    l.resource_limit
FROM `limit` l
INNER JOIN registered_limit rl ON l.registered_limit_id = rl.id
INNER JOIN service s ON rl.service_id = s.id
WHERE s.type = 'compute'
  AND l.project_id IS NOT NULL
  AND (rl.region_id = sqlc.arg(region_id) OR (rl.region_id IS NULL AND sqlc.arg(region_id) = ''))