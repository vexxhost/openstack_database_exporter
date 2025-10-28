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
