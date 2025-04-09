use crate::{
    database::{DatabaseCollector, Pool},
    neutron::schema::{agents, ha_router_agent_port_bindings},
};
use chrono::prelude::*;
use diesel::prelude::*;
use prometheus::{
    IntGaugeVec, Opts,
    core::{Collector, Desc},
    proto::MetricFamily,
};
use tracing::{Level, error, span};

const METRICS_NUMBER: usize = 1;

#[derive(Queryable, Insertable, Selectable, Debug)]
#[diesel(check_for_backend(diesel::sqlite::Sqlite))]
#[diesel(check_for_backend(diesel::mysql::Mysql))]
#[diesel(table_name = ha_router_agent_port_bindings)]
struct HaRouterAgentPortBinding {
    port_id: String,
    router_id: String,
    l3_agent_id: Option<String>,
    state: String,
}

#[derive(Queryable, Insertable, Selectable, Debug)]
#[diesel(check_for_backend(diesel::sqlite::Sqlite))]
#[diesel(check_for_backend(diesel::mysql::Mysql))]
#[diesel(table_name = agents)]
struct Agent {
    id: String,
    host: String,
    admin_state_up: bool,
    heartbeat_timestamp: NaiveDateTime,
}

pub struct NeutronCollector {
    pool: Pool,

    l3_agent_of_router: IntGaugeVec,
}

impl DatabaseCollector for NeutronCollector {
    fn new(pool: Pool) -> Self {
        let l3_agent_of_router = IntGaugeVec::new(
            Opts::new(
                "openstack_neutron_l3_agent_of_router".to_owned(),
                "l3_agent_of_router".to_owned(),
            ),
            &[
                "router_id",
                "l3_agent_id",
                "ha_state",
                "agent_alive",
                "agent_admin_up",
                "agent_host",
            ],
        )
        .unwrap();

        Self {
            pool,

            l3_agent_of_router,
        }
    }
}

impl Collector for NeutronCollector {
    fn desc(&self) -> Vec<&Desc> {
        let mut desc = Vec::with_capacity(METRICS_NUMBER);
        desc.extend(self.l3_agent_of_router.desc());

        desc
    }

    fn collect(&self) -> Vec<MetricFamily> {
        use crate::neutron::schema::ha_router_agent_port_bindings::dsl::*;

        let span = span!(Level::INFO, "neutron_collector");
        let _enter = span.enter();

        let mut mfs = Vec::with_capacity(METRICS_NUMBER);

        let mut conn = match self.pool.get() {
            Ok(conn) => conn,
            Err(err) => {
                error!(error = %err, "failed to get connection from pool");
                return mfs;
            }
        };

        let data = match ha_router_agent_port_bindings
            .inner_join(agents::table)
            .select((HaRouterAgentPortBinding::as_select(), Agent::as_select()))
            .load::<(HaRouterAgentPortBinding, Agent)>(&mut conn)
        {
            Ok(data) => data,
            Err(err) => {
                error!(error = %err, "failed to load data");
                return mfs;
            }
        };

        self.l3_agent_of_router.reset();

        data.iter()
            .for_each(|(ha_router_agent_port_binding, agent)| {
                let alive = agent
                    .heartbeat_timestamp
                    .signed_duration_since(Utc::now().naive_utc())
                    .num_seconds()
                    < 75;

                self.l3_agent_of_router
                    .with_label_values(&[
                        ha_router_agent_port_binding.router_id.clone(),
                        agent.id.clone(),
                        ha_router_agent_port_binding.state.clone(),
                        alive.to_string(),
                        agent.admin_state_up.to_string(),
                        agent.host.clone(),
                    ])
                    .set(if alive { 1 } else { 0 });
            });

        mfs.extend(self.l3_agent_of_router.collect());
        mfs
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::database;
    use indoc::indoc;
    use pretty_assertions::assert_eq;
    use prometheus::TextEncoder;

    #[test]
    fn test_collector() {
        use crate::neutron::schema::{agents::dsl::*, ha_router_agent_port_bindings::dsl::*};

        let pool = database::connect(":memory:").unwrap();
        diesel::sql_query(indoc! {r#"
            CREATE TABLE agents (
                `id` varchar(36) NOT NULL,
                `host` varchar(255) NOT NULL,
                `admin_state_up` tinyint(1) NOT NULL DEFAULT 1,
                `heartbeat_timestamp` datetime NOT NULL
            )"#})
        .execute(&mut pool.get().unwrap())
        .unwrap();
        diesel::insert_into(agents)
            .values(&Agent {
                id: "ddbf087c-e38f-4a73-bcb3-c38f2a719a03".into(),
                host: "dev-os-ctrl-02".into(),
                admin_state_up: true,
                heartbeat_timestamp: Utc::now().naive_utc(),
            })
            .execute(&mut pool.get().unwrap())
            .unwrap();
        diesel::sql_query(indoc! {r#"
            CREATE TABLE ha_router_agent_port_bindings (
                `port_id` varchar(36) NOT NULL,
                `router_id` varchar(36) NOT NULL,
                `l3_agent_id` varchar(36) DEFAULT NULL,
                `state` varchar(36) DEFAULT NULL
            )"#})
        .execute(&mut pool.get().unwrap())
        .unwrap();
        diesel::insert_into(ha_router_agent_port_bindings)
            .values(&HaRouterAgentPortBinding {
                port_id: "f8a44de0-fc8e-45df-93c7-f79bf3b01c95".into(),
                router_id: "9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f".into(),
                l3_agent_id: Some("ddbf087c-e38f-4a73-bcb3-c38f2a719a03".into()),
                state: "active".into(),
            })
            .execute(&mut pool.get().unwrap())
            .unwrap();
        diesel::insert_into(ha_router_agent_port_bindings)
            .values(&HaRouterAgentPortBinding {
                port_id: "9135549f-07b7-4efd-9a81-3b71fc69b8f8".into(),
                router_id: "f8a44de0-fc8e-45df-93c7-f79bf3b01c95".into(),
                l3_agent_id: Some("ddbf087c-e38f-4a73-bcb3-c38f2a719a03".into()),
                state: "backup".into(),
            })
            .execute(&mut pool.get().unwrap())
            .unwrap();

        let collector = NeutronCollector::new(pool.clone());

        let mfs = collector.collect();
        assert_eq!(mfs.len(), METRICS_NUMBER);

        let encoder = TextEncoder::new();
        assert_eq!(
            encoder.encode_to_string(&mfs).unwrap(),
            indoc! {r#"
                # HELP openstack_neutron_l3_agent_of_router l3_agent_of_router
                # TYPE openstack_neutron_l3_agent_of_router gauge
                openstack_neutron_l3_agent_of_router{agent_admin_up="true",agent_alive="true",agent_host="dev-os-ctrl-02",ha_state="backup",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="f8a44de0-fc8e-45df-93c7-f79bf3b01c95"} 1
                openstack_neutron_l3_agent_of_router{agent_admin_up="true",agent_alive="true",agent_host="dev-os-ctrl-02",ha_state="active",l3_agent_id="ddbf087c-e38f-4a73-bcb3-c38f2a719a03",router_id="9daeb7dd-7e3f-4e44-8c42-c7a0e8c8a42f"} 1
                "#
            }
        );
    }
}
