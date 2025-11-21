-- name: GetShareMetrics :many
-- Get share metrics for openstack_sharev2_share_gb and openstack_sharev2_share_status
-- This joins shares with share_instances to get current status and availability zone info
SELECT 
    s.id,
    s.display_name as name,
    s.project_id,
    s.size,
    s.share_proto,
    si.status,
    COALESCE(st.name, '') as share_type_name,
    COALESCE(az.name, '') as availability_zone
FROM shares s
LEFT JOIN share_instances si ON s.id = si.share_id AND si.deleted = 'False'
LEFT JOIN share_types st ON si.share_type_id = st.id AND st.deleted = 'False'  
LEFT JOIN availability_zones az ON si.availability_zone_id = az.id AND az.deleted = 'False'
WHERE s.deleted = 'False'
ORDER BY s.created_at;
