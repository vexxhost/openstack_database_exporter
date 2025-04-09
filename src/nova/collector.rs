use crate::{
    database::{DatabaseCollector, Pool},
    nova::schema::instances,
};
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
#[diesel(table_name = instances)]
struct Instance {
    uuid: String,
    task_state: Option<String>,
}

pub struct NovaCollector {
    pool: Pool,

    server_task_state: IntGaugeVec,
}

impl DatabaseCollector for NovaCollector {
    fn new(pool: Pool) -> Self {
        let server_task_state = IntGaugeVec::new(
            Opts::new(
                "openstack_nova_server_task_state".to_owned(),
                "server_task_state".to_owned(),
            ),
            &["id", "task_state"],
        )
        .unwrap();

        Self {
            pool,

            server_task_state,
        }
    }
}

impl Collector for NovaCollector {
    fn desc(&self) -> Vec<&Desc> {
        let mut desc = Vec::with_capacity(METRICS_NUMBER);
        desc.extend(self.server_task_state.desc());

        desc
    }

    fn collect(&self) -> Vec<MetricFamily> {
        use crate::nova::schema::instances::dsl::*;

        let span = span!(Level::INFO, "nova_collector");
        let _enter = span.enter();

        let mut mfs = Vec::with_capacity(METRICS_NUMBER);

        let mut conn = match self.pool.get() {
            Ok(conn) => conn,
            Err(err) => {
                error!(error = %err, "failed to get connection from pool");
                return mfs;
            }
        };

        let data = match instances
            .filter(deleted.eq(0))
            .select(Instance::as_select())
            .load::<Instance>(&mut conn)
        {
            Ok(data) => data,
            Err(err) => {
                error!(error = %err, "failed to load data");
                return mfs;
            }
        };

        self.server_task_state.reset();

        data.iter().for_each(|instance| {
            let task_state_value = if instance.task_state.is_none() { 0 } else { 1 };
            self.server_task_state
                .with_label_values(&[
                    instance.uuid.clone(),
                    instance.task_state.clone().unwrap_or_default(),
                ])
                .set(task_state_value);
        });

        mfs.extend(self.server_task_state.collect());
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
        use crate::nova::schema::instances::dsl::*;

        let pool = database::connect(":memory:").unwrap();
        diesel::sql_query(indoc! {r#"
            CREATE TABLE instances (
                `id` INTEGER PRIMARY KEY,
                `uuid` varchar(36) NOT NULL,
                `task_state` varchar(255) DEFAULT NULL,
                `deleted` int DEFAULT 0
            )"#})
        .execute(&mut pool.get().unwrap())
        .unwrap();
        diesel::insert_into(instances)
            .values(&Instance {
                uuid: "ec2917d8-cbd4-49b2-b204-f2c0a81cbe3b".to_string(),
                task_state: None,
            })
            .execute(&mut pool.get().unwrap())
            .unwrap();
        diesel::insert_into(instances)
            .values(&Instance {
                uuid: "f3e2e9b6-3b7d-4b1e-9e0d-0f6b3b3b1b1b".to_string(),
                task_state: Some("spawning".into()),
            })
            .execute(&mut pool.get().unwrap())
            .unwrap();

        let collector = NovaCollector::new(pool.clone());

        let mfs = collector.collect();
        assert_eq!(mfs.len(), METRICS_NUMBER);

        let encoder = TextEncoder::new();
        assert_eq!(
            encoder.encode_to_string(&mfs).unwrap(),
            indoc! {r#"
                # HELP openstack_nova_server_task_state server_task_state
                # TYPE openstack_nova_server_task_state gauge
                openstack_nova_server_task_state{id="f3e2e9b6-3b7d-4b1e-9e0d-0f6b3b3b1b1b",task_state="spawning"} 1
                openstack_nova_server_task_state{id="ec2917d8-cbd4-49b2-b204-f2c0a81cbe3b",task_state=""} 0
                "#
            }
        );
    }
}
