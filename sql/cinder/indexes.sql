-- Indexes for optimizing Cinder queries

-- For the quotas query
CREATE INDEX IF NOT EXISTS ix_quotas_deleted_resource ON quotas(deleted, resource);

-- For the volumes query (USE INDEX hint in query requires this)
CREATE INDEX IF NOT EXISTS volumes_service_uuid_idx ON volumes(service_uuid);
CREATE INDEX IF NOT EXISTS ix_volumes_deleted ON volumes(deleted);
CREATE INDEX IF NOT EXISTS ix_volume_types_id_deleted ON volume_types(id, deleted);
CREATE INDEX IF NOT EXISTS ix_volume_attachment_volume_id_deleted ON volume_attachment(volume_id, deleted);

-- These indexes should already exist but listing for completeness
-- CREATE INDEX ix_volumes_project_id ON volumes(project_id);
-- CREATE INDEX ix_quotas_project_id ON quotas(project_id);
-- CREATE INDEX ix_quota_usages_project_id ON quota_usages(project_id);