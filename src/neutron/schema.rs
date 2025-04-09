use diesel::{allow_tables_to_appear_in_same_query, joinable, table};

table! {
    ha_router_agent_port_bindings (port_id) {
        #[max_length = 36]
        port_id -> Varchar,
        #[max_length = 36]
        router_id -> Varchar,
        #[max_length = 36]
        l3_agent_id -> Nullable<Varchar>,
        // NOTE(mnaser): diesel_derive_enum has issues with MultiConnection
        //               https://github.com/adwhit/diesel-derive-enum/issues/105
        state -> Varchar,
    }
}

table! {
    agents (id) {
        #[max_length = 36]
        id -> Varchar,
        #[max_length = 255]
        host -> Varchar,
        admin_state_up -> Bool,
        heartbeat_timestamp -> Timestamp
    }
}

joinable!(ha_router_agent_port_bindings -> agents (l3_agent_id));
allow_tables_to_appear_in_same_query!(agents, ha_router_agent_port_bindings);
