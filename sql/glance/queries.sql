-- name: GetAllImages :many
SELECT
    id,
    name,
    size,
    status,
    owner,
    visibility,
    disk_format,
    container_format,
    checksum,
    created_at,
    updated_at,
    min_disk,
    min_ram,
    protected,
    virtual_size,
    os_hidden,
    os_hash_algo,
    os_hash_value
FROM
    images
WHERE
    deleted = 0;

-- name: GetImageCount :one
SELECT
    COUNT(*) as count
FROM
    images
WHERE
    deleted = 0;