-- Nova API database indexes for performance optimization

-- Flavors indexes
CREATE INDEX IF NOT EXISTS flavors_disabled_idx ON flavors(disabled);
CREATE INDEX IF NOT EXISTS flavors_is_public_idx ON flavors(is_public);

-- Quotas indexes  
CREATE INDEX IF NOT EXISTS quotas_resource_idx ON quotas(resource);
CREATE INDEX IF NOT EXISTS quotas_hard_limit_idx ON quotas(hard_limit);

-- Aggregates indexes
CREATE INDEX IF NOT EXISTS aggregates_name_idx ON aggregates(name);

-- Aggregate hosts indexes
CREATE INDEX IF NOT EXISTS aggregate_hosts_host_idx ON aggregate_hosts(host);
