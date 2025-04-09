use crate::{
    database::{DatabaseCollector, Pool},
    nova_api::schema::build_requests,
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
#[diesel(table_name = build_requests)]
struct BuildRequest {
    project_id: String,
    instance_uuid: Option<String>,
}

pub struct NovaApiCollector {
    pool: Pool,

    build_request: IntGaugeVec,
}

impl DatabaseCollector for NovaApiCollector {
    fn new(pool: Pool) -> Self {
        let build_request = IntGaugeVec::new(
            Opts::new(
                "openstack_nova_api_build_request".to_owned(),
                "build_request".to_owned(),
            ),
            &["project_id", "instance_uuid"],
        )
        .unwrap();

        Self {
            pool,

            build_request,
        }
    }
}

impl Collector for NovaApiCollector {
    fn desc(&self) -> Vec<&Desc> {
        let mut desc = Vec::with_capacity(METRICS_NUMBER);
        desc.extend(self.build_request.desc());

        desc
    }

    fn collect(&self) -> Vec<MetricFamily> {
        use crate::nova_api::schema::build_requests::dsl::*;

        let span = span!(Level::INFO, "nova_api_collector");
        let _enter = span.enter();

        let mut mfs = Vec::with_capacity(METRICS_NUMBER);

        let mut conn = match self.pool.get() {
            Ok(conn) => conn,
            Err(err) => {
                error!(error = %err, "failed to get connection from pool");
                return mfs;
            }
        };

        let data = match build_requests
            .select(BuildRequest::as_select())
            .load::<BuildRequest>(&mut conn)
        {
            Ok(data) => data,
            Err(err) => {
                error!(error = %err, "failed to load data");
                return mfs;
            }
        };

        self.build_request.reset();

        data.iter().for_each(|build_request| {
            let uuid = match &build_request.instance_uuid {
                Some(uuid) => uuid,
                None => return,
            };

            self.build_request
                .with_label_values(&[build_request.project_id.clone(), uuid.into()])
                .set(1);
        });

        mfs.extend(self.build_request.collect());
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
        use crate::nova_api::schema::build_requests::dsl::*;

        let pool = database::connect(":memory:").unwrap();
        diesel::sql_query(indoc! {r#"
            CREATE TABLE build_requests (
                `id` INTEGER PRIMARY KEY,
                `project_id` varchar(255) NOT NULL,
                `instance_uuid` varchar(36) DEFAULT NULL
            )"#})
        .execute(&mut pool.get().unwrap())
        .unwrap();
        diesel::insert_into(build_requests)
            .values(&BuildRequest {
                project_id: "ec2917d8-cbd4-49b2-b204-f2c0a81cbe3b".into(),
                instance_uuid: Some("f3e2e9b6-3b7d-4b1e-9e0d-0f6b3b3b1b1b".into()),
            })
            .execute(&mut pool.get().unwrap())
            .unwrap();
        diesel::insert_into(build_requests)
            .values(&BuildRequest {
                project_id: "107b88ab-f104-4ac5-8032-302e8a621d46".into(),
                instance_uuid: Some("894cacd1-a432-4093-a0e7-cd29503205da".into()),
            })
            .execute(&mut pool.get().unwrap())
            .unwrap();

        let collector = NovaApiCollector::new(pool.clone());

        let mfs = collector.collect();
        assert_eq!(mfs.len(), METRICS_NUMBER);

        let encoder = TextEncoder::new();
        assert_eq!(
            encoder.encode_to_string(&mfs).unwrap(),
            indoc! {r#"
                # HELP openstack_nova_api_build_request build_request
                # TYPE openstack_nova_api_build_request gauge
                openstack_nova_api_build_request{instance_uuid="f3e2e9b6-3b7d-4b1e-9e0d-0f6b3b3b1b1b",project_id="ec2917d8-cbd4-49b2-b204-f2c0a81cbe3b"} 1
                openstack_nova_api_build_request{instance_uuid="894cacd1-a432-4093-a0e7-cd29503205da",project_id="107b88ab-f104-4ac5-8032-302e8a621d46"} 1
                "#
            }
        );
    }
}
