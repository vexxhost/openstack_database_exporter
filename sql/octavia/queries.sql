-- name: GetAllPools :many
SELECT
    id,
    project_id,
    name,
    protocol,
    lb_algorithm,
    operating_status,
    load_balancer_id,
    provisioning_status
FROM
    pool;

-- name: GetAllLoadBalancersWithVip :many
SELECT
    lb.id,
    lb.project_id,
    lb.name,
    lb.provisioning_status,
    lb.operating_status,
    lb.provider,
    v.ip_address as vip_address
FROM
    load_balancer lb
    LEFT JOIN vip v ON lb.id = v.load_balancer_id;

-- name: GetAllAmphora :many
SELECT
    id,
    compute_id,
    status,
    load_balancer_id,
    lb_network_ip,
    ha_ip,
    role,
    cert_expiration
FROM
    amphora;
