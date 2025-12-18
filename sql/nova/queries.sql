-- name: GetInstances :many
SELECT 
    id,
    uuid,
    display_name,
    user_id,
    project_id,
    host,
    availability_zone,
    vm_state,
    power_state,
    task_state,
    memory_mb,
    vcpus,
    root_gb,
    ephemeral_gb,
    launched_at,
    terminated_at,
    instance_type_id,
    deleted
FROM instances
WHERE deleted = 0;

-- name: GetServices :many
SELECT 
    id,
    uuid,
    host,
    `binary`,
    topic,
    disabled,
    disabled_reason,
    last_seen_up,
    forced_down,
    version,
    report_count,
    deleted
FROM services
WHERE deleted = 0;

-- name: GetComputeNodes :many
SELECT 
    id,
    uuid,
    host,
    hypervisor_hostname,
    hypervisor_type,
    hypervisor_version,
    vcpus,
    vcpus_used,
    memory_mb,
    memory_mb_used,
    local_gb,
    local_gb_used,
    disk_available_least,
    free_ram_mb,
    free_disk_gb,
    current_workload,
    running_vms,
    cpu_allocation_ratio,
    ram_allocation_ratio,
    disk_allocation_ratio,
    deleted
FROM compute_nodes
WHERE deleted = 0;
