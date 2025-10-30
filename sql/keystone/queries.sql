-- name: GetProjectMetrics :many
SELECT 
    p.id,
    p.name,
    COALESCE(p.description, '') as description,
    p.enabled,
    p.domain_id,
    COALESCE(p.parent_id, '') as parent_id,
    p.is_domain,
    COALESCE(GROUP_CONCAT(pt.name SEPARATOR ','), '') as tags
FROM project p
LEFT JOIN project_tag pt ON p.id = pt.project_id
WHERE p.is_domain = 0
GROUP BY p.id, p.name, p.description, p.enabled, p.domain_id, p.parent_id, p.is_domain;

-- name: GetDomainMetrics :many
SELECT 
    id,
    name,
    COALESCE(description, '') as description,
    enabled
FROM project 
WHERE is_domain = 1 AND id != '<<keystone.domain.root>>';

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