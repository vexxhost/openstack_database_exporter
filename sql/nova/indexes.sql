-- Nova database indexes for performance optimization

-- Instances indexes
CREATE INDEX IF NOT EXISTS instances_vm_state_idx ON instances(vm_state);
CREATE INDEX IF NOT EXISTS instances_power_state_idx ON instances(power_state);  
CREATE INDEX IF NOT EXISTS instances_task_state_idx ON instances(task_state);
CREATE INDEX IF NOT EXISTS instances_host_idx ON instances(host);

-- Services indexes  
CREATE INDEX IF NOT EXISTS services_binary_idx ON services(`binary`);
CREATE INDEX IF NOT EXISTS services_disabled_idx ON services(disabled);
CREATE INDEX IF NOT EXISTS services_last_seen_up_idx ON services(last_seen_up);

-- Compute nodes indexes
CREATE INDEX IF NOT EXISTS compute_nodes_host_idx ON compute_nodes(host);
CREATE INDEX IF NOT EXISTS compute_nodes_hypervisor_hostname_idx ON compute_nodes(hypervisor_hostname);
