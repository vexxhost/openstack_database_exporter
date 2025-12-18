-- name: GetResourceMetrics :many
-- This is the main query that provides data for all four metrics:
-- - resource_total: inventory total
-- - resource_allocation_ratio: inventory allocation_ratio  
-- - resource_reserved: inventory reserved
-- - resource_usage: sum of allocations per resource provider + class
SELECT 
    rp.name as hostname,
    rc.name as resource_type,
    i.total,
    i.allocation_ratio,
    i.reserved,
    COALESCE(SUM(a.used), 0) as used
FROM resource_providers rp
JOIN inventories i ON rp.id = i.resource_provider_id
JOIN resource_classes rc ON i.resource_class_id = rc.id
LEFT JOIN allocations a ON rp.id = a.resource_provider_id AND rc.id = a.resource_class_id
GROUP BY rp.id, rp.name, rc.id, rc.name, i.total, i.allocation_ratio, i.reserved
ORDER BY rp.name, rc.name;

-- name: GetAllocationsByProject :many
-- Get resource usage by project for Nova quota calculations
SELECT 
    p.external_id as project_id,
    rc.name as resource_type,
    COALESCE(SUM(a.used), 0) as used
FROM projects p
LEFT JOIN consumers c ON p.id = c.project_id
LEFT JOIN allocations a ON c.uuid = a.consumer_id
LEFT JOIN resource_classes rc ON a.resource_class_id = rc.id
WHERE rc.name IS NOT NULL
GROUP BY p.external_id, rc.name
ORDER BY p.external_id, rc.name;

-- name: GetResourceClasses :many
-- Get all resource classes for reference
SELECT 
    id,
    name
FROM resource_classes
ORDER BY name;

-- name: GetConsumers :many
-- Get consumer information for allocation tracking  
SELECT 
    c.id,
    c.uuid,
    c.generation,
    p.external_id as project_id,
    u.external_id as user_id
FROM consumers c
JOIN projects p ON c.project_id = p.id
JOIN users u ON c.user_id = u.id
ORDER BY c.created_at DESC;
