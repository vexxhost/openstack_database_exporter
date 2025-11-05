-- name: GetClusterMetrics :many
SELECT 
    c.uuid,
    c.name,
    COALESCE(c.stack_id, '') as stack_id,
    COALESCE(c.status, '') as status,
    c.project_id,
    COALESCE(master_ng.node_count, 0) as master_count,
    COALESCE(worker_ng.node_count, 0) as node_count
FROM cluster c
LEFT JOIN (
    SELECT cluster_id, SUM(node_count) as node_count
    FROM nodegroup 
    WHERE role = 'master'
    GROUP BY cluster_id
) master_ng ON c.uuid = master_ng.cluster_id
LEFT JOIN (
    SELECT cluster_id, SUM(node_count) as node_count
    FROM nodegroup 
    WHERE role = 'worker'
    GROUP BY cluster_id
) worker_ng ON c.uuid = worker_ng.cluster_id;