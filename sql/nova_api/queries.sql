-- name: GetFlavors :many
SELECT 
    id,
    flavorid,
    name,
    vcpus,
    memory_mb,
    root_gb,
    ephemeral_gb,
    swap,
    rxtx_factor,
    disabled,
    is_public
FROM flavors;

-- name: GetQuotas :many
SELECT 
    id,
    project_id,
    resource,
    hard_limit
FROM quotas;

-- name: GetAggregates :many
SELECT 
    id,
    uuid,
    name,
    created_at,
    updated_at
FROM aggregates;

-- name: GetAggregateHosts :many
SELECT 
    ah.id,
    ah.host,
    ah.aggregate_id,
    a.name as aggregate_name,
    a.uuid as aggregate_uuid
FROM aggregate_hosts ah
JOIN aggregates a ON ah.aggregate_id = a.id;

-- name: GetQuotaUsages :many
SELECT 
    id,
    project_id,
    resource,
    in_use,
    reserved,
    until_refresh,
    user_id
FROM quota_usages;
