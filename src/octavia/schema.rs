use diesel::{allow_tables_to_appear_in_same_query, joinable, table};

table! {
    load_balancer (id) {
        #[max_length = 36]
        project_id -> Nullable<Varchar>,
        #[max_length = 36]
        id -> Varchar,
        #[max_length = 255]
        name -> Nullable<Varchar>,
        #[max_length = 16]
        provisioning_status -> Varchar,
        #[max_length = 16]
        operating_status -> Varchar,
        #[max_length = 64]
        provider -> Nullable<Varchar>,
    }
}

table! {
    vip (load_balancer_id) {
        #[max_length = 36]
        load_balancer_id -> Varchar,
        #[max_length = 64]
        ip_address -> Nullable<Varchar>,
    }
}

table! {
    amphora (id) {
        #[max_length = 36]
        id -> Varchar,
        #[max_length = 36]
        compute_id -> Nullable<Varchar>,
        #[max_length = 36]
        status -> Varchar,
        #[max_length = 36]
        load_balancer_id -> Nullable<Varchar>,
        #[max_length = 64]
        lb_network_ip -> Nullable<Varchar>,
        #[max_length = 64]
        ha_ip -> Nullable<Varchar>,
        #[max_length = 36]
        role -> Nullable<Varchar>,
        cert_expiration -> Nullable<Timestamp>,
    }
}

table! {
    pool (id) {
        #[max_length = 36]
        project_id -> Nullable<Varchar>,
        #[max_length = 36]
        id -> Varchar,
        #[max_length = 255]
        name -> Nullable<Varchar>,
        #[max_length = 16]
        protocol -> Varchar,
        #[max_length = 255]
        lb_algorithm -> Varchar,
        #[max_length = 16]
        operating_status -> Varchar,
        #[max_length = 36]
        load_balancer_id -> Nullable<Varchar>,
        #[max_length = 16]
        provisioning_status -> Varchar,
    }
}

joinable!(vip -> load_balancer (load_balancer_id));
allow_tables_to_appear_in_same_query!(amphora, load_balancer, pool, vip);
