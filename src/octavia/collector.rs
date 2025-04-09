use crate::{
    database::{AnyConnection, DatabaseCollector, Pool as DatabasePool},
    octavia::schema::{amphora, load_balancer, pool, vip},
};
use chrono::SecondsFormat;
use diesel::{prelude::*, r2d2::ConnectionManager};
use prometheus::{
    IntGauge, IntGaugeVec, Opts,
    core::{Collector, Desc},
    proto::MetricFamily,
};
use r2d2::PooledConnection;
use tracing::{Level, error, span};

const METRICS_NUMBER: usize = 7;

#[derive(Queryable, Insertable, Selectable, Debug)]
#[diesel(check_for_backend(diesel::sqlite::Sqlite))]
#[diesel(check_for_backend(diesel::mysql::Mysql))]
#[diesel(table_name = load_balancer)]
struct LoadBalancer {
    project_id: Option<String>,
    id: String,
    name: Option<String>,
    provisioning_status: String,
    operating_status: String,
    provider: Option<String>,
}

#[derive(Queryable, Insertable, Selectable, Debug)]
#[diesel(check_for_backend(diesel::sqlite::Sqlite))]
#[diesel(check_for_backend(diesel::mysql::Mysql))]
#[diesel(table_name = vip)]
struct VirtualIp {
    load_balancer_id: String,
    ip_address: Option<String>,
}

#[derive(Queryable, Insertable, Selectable, Debug)]
#[diesel(check_for_backend(diesel::sqlite::Sqlite))]
#[diesel(check_for_backend(diesel::mysql::Mysql))]
#[diesel(table_name = amphora)]
struct Amphora {
    id: String,
    compute_id: Option<String>,
    status: String,
    load_balancer_id: Option<String>,
    lb_network_ip: Option<String>,
    ha_ip: Option<String>,
    role: Option<String>,
    cert_expiration: Option<chrono::NaiveDateTime>,
}

#[derive(Queryable, Insertable, Selectable, Debug)]
#[diesel(check_for_backend(diesel::sqlite::Sqlite))]
#[diesel(check_for_backend(diesel::mysql::Mysql))]
#[diesel(table_name = pool)]
struct Pool {
    project_id: Option<String>,
    id: String,
    name: Option<String>,
    protocol: String,
    lb_algorithm: String,
    operating_status: String,
    load_balancer_id: Option<String>,
    provisioning_status: String,
}

pub struct OctaviaCollector {
    pool: DatabasePool,

    pool_status: IntGaugeVec,
    loadbalancer_status: IntGaugeVec,
    amphora_status: IntGaugeVec,
    total_amphorae: IntGauge,
    total_loadbalancers: IntGauge,
    total_pools: IntGauge,
    up: IntGauge,
}

impl DatabaseCollector for OctaviaCollector {
    fn new(pool: DatabasePool) -> Self {
        let pool_status = IntGaugeVec::new(
            Opts::new(
                "openstack_loadbalancer_pool_status".to_string(),
                "pool_status".to_string(),
            ),
            &[
                "id",
                "provisioning_status",
                "name",
                "loadbalancers",
                "protocol",
                "lb_algorithm",
                "operating_status",
                "project_id",
            ],
        )
        .unwrap();
        let loadbalancer_status = IntGaugeVec::new(
            Opts::new(
                "openstack_loadbalancer_loadbalancer_status".to_string(),
                "loadbalancer_status".to_string(),
            ),
            &[
                "id",
                "name",
                "project_id",
                "operating_status",
                "provisioning_status",
                "provider",
                "vip_address",
            ],
        )
        .unwrap();
        let amphora_status = IntGaugeVec::new(
            Opts::new(
                "openstack_loadbalancer_amphora_status".to_string(),
                "amphora_status".to_string(),
            ),
            &[
                "id",
                "loadbalancer_id",
                "compute_id",
                "status",
                "role",
                "lb_network_ip",
                "ha_ip",
                "cert_expiration",
            ],
        )
        .unwrap();
        let total_amphorae = IntGauge::with_opts(Opts::new(
            "openstack_loadbalancer_total_amphorae".to_string(),
            "total_amphorae".to_string(),
        ))
        .unwrap();
        let total_loadbalancers = IntGauge::with_opts(Opts::new(
            "openstack_loadbalancer_total_loadbalancers".to_string(),
            "total_loadbalancers".to_string(),
        ))
        .unwrap();
        let total_pools = IntGauge::with_opts(Opts::new(
            "openstack_loadbalancer_total_pools".to_string(),
            "total_pools".to_string(),
        ))
        .unwrap();

        let up = IntGauge::with_opts(Opts::new(
            "openstack_loadbalancer_up".to_string(),
            "up".to_string(),
        ))
        .unwrap();
        up.set(1);

        Self {
            pool,

            pool_status,
            loadbalancer_status,
            amphora_status,
            total_amphorae,
            total_loadbalancers,
            total_pools,
            up,
        }
    }
}

impl Collector for OctaviaCollector {
    fn desc(&self) -> Vec<&Desc> {
        let mut desc = Vec::with_capacity(METRICS_NUMBER);
        desc.extend(self.loadbalancer_status.desc());

        desc
    }

    fn collect(&self) -> Vec<MetricFamily> {
        let span = span!(Level::INFO, "octavia_collector");
        let _enter = span.enter();

        let mut mfs = Vec::with_capacity(METRICS_NUMBER);

        let mut conn = match self.pool.get() {
            Ok(conn) => conn,
            Err(err) => {
                error!(error = %err, "failed to get connection from pool");
                return mfs;
            }
        };

        self.collect_pool_status(&mut conn);
        self.collect_amphora_status(&mut conn);
        self.collect_loadbalancer_status(&mut conn);

        mfs.extend(self.pool_status.collect());
        mfs.extend(self.amphora_status.collect());
        mfs.extend(self.loadbalancer_status.collect());
        mfs.extend(self.total_amphorae.collect());
        mfs.extend(self.total_loadbalancers.collect());
        mfs.extend(self.total_pools.collect());
        mfs.extend(self.up.collect());

        mfs
    }
}

impl OctaviaCollector {
    fn collect_pool_status(&self, conn: &mut PooledConnection<ConnectionManager<AnyConnection>>) {
        use crate::octavia::schema::pool::dsl::*;

        let data = match pool.select(Pool::as_select()).load::<Pool>(conn) {
            Ok(data) => data,
            Err(err) => {
                error!(error = %err, "failed to load data");
                return;
            }
        };

        self.pool_status.reset();

        data.iter().for_each(|pool_inst| {
            let value = match pool_inst.provisioning_status.as_str() {
                "ACTIVE" => 0,
                "DELETED" => 1,
                "ERROR" => 2,
                "PENDING_CREATE" => 3,
                "PENDING_UPDATE" => 4,
                "PENDING_DELETE" => 5,
                _ => -1,
            };

            self.pool_status
                .with_label_values(&[
                    pool_inst.id.clone(),
                    pool_inst.provisioning_status.clone(),
                    pool_inst.name.clone().unwrap_or_default(),
                    pool_inst.load_balancer_id.clone().unwrap_or_default(),
                    pool_inst.protocol.clone(),
                    pool_inst.lb_algorithm.clone(),
                    pool_inst.operating_status.clone(),
                    pool_inst.project_id.clone().unwrap_or_default(),
                ])
                .set(value);
        });

        self.total_pools.set(data.len() as i64);
    }

    fn collect_amphora_status(
        &self,
        conn: &mut PooledConnection<ConnectionManager<AnyConnection>>,
    ) {
        use crate::octavia::schema::amphora::dsl::*;

        let data = match amphora.select(Amphora::as_select()).load::<Amphora>(conn) {
            Ok(data) => data,
            Err(err) => {
                error!(error = %err, "failed to load data");
                return;
            }
        };

        self.loadbalancer_status.reset();

        data.iter().for_each(|amphora_inst| {
            let value = match amphora_inst.status.as_str() {
                "BOOTING" => 0,
                "ALLOCATED" => 1,
                "READY" => 2,
                "PENDING_CREATE" => 3,
                "PENDING_DELETE" => 4,
                "DELETED" => 5,
                "ERROR" => 6,
                _ => -1,
            };

            self.amphora_status
                .with_label_values(&[
                    amphora_inst.id.clone(),
                    amphora_inst.load_balancer_id.clone().unwrap_or_default(),
                    amphora_inst.compute_id.clone().unwrap_or_default(),
                    amphora_inst.status.clone(),
                    amphora_inst.role.clone().unwrap_or_default(),
                    amphora_inst.lb_network_ip.clone().unwrap_or_default(),
                    amphora_inst.ha_ip.clone().unwrap_or_default(),
                    amphora_inst
                        .cert_expiration
                        .unwrap_or_default()
                        .and_utc()
                        .to_rfc3339_opts(SecondsFormat::Secs, true),
                ])
                .set(value);
        });

        self.total_amphorae.set(data.len() as i64);
    }

    fn collect_loadbalancer_status(
        &self,
        conn: &mut PooledConnection<ConnectionManager<AnyConnection>>,
    ) {
        use crate::octavia::schema::load_balancer::dsl::*;

        let data = match load_balancer
            .inner_join(vip::table)
            .select((LoadBalancer::as_select(), VirtualIp::as_select()))
            .load::<(LoadBalancer, VirtualIp)>(conn)
        {
            Ok(data) => data,
            Err(err) => {
                error!(error = %err, "failed to load data");
                return;
            }
        };

        self.loadbalancer_status.reset();

        data.iter().for_each(|(loadbalancer, vip)| {
            let value = match loadbalancer.operating_status.as_str() {
                "ONLINE" => 0,
                "DRAINING" => 1,
                "OFFLINE" => 2,
                "ERROR" => 3,
                "NO_MONITOR" => 4,
                _ => -1,
            };

            self.loadbalancer_status
                .with_label_values(&[
                    loadbalancer.id.clone(),
                    loadbalancer.name.clone().unwrap_or_default(),
                    loadbalancer.project_id.clone().unwrap_or_default(),
                    loadbalancer.operating_status.clone(),
                    loadbalancer.provisioning_status.clone(),
                    loadbalancer.provider.clone().unwrap_or_default(),
                    vip.ip_address.clone().unwrap_or_default(),
                ])
                .set(value);
        });

        self.total_loadbalancers.set(data.len() as i64);
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
        use crate::octavia::schema::{
            amphora::dsl::*, load_balancer::dsl::*, pool::dsl::*, vip::dsl::*,
        };

        let database_pool = database::connect(":memory:").unwrap();

        diesel::sql_query(indoc! {r#"
            CREATE TABLE load_balancer (
                `project_id` varchar(36) NOT NULL,
                `id` varchar(36) NOT NULL,
                `name` varchar(255),
                `provisioning_status` varchar(16) NOT NULL,
                `operating_status` varchar(16) NOT NULL,
                `provider` varchar(64)
            )"#})
        .execute(&mut database_pool.get().unwrap())
        .unwrap();
        diesel::insert_into(load_balancer)
            .values(&LoadBalancer {
                project_id: Some("e3cd678b11784734bc366148aa37580e".into()),
                id: "607226db-27ef-4d41-ae89-f2a800e9c2db".into(),
                name: Some("best_load_balancer".into()),
                provisioning_status: "ACTIVE".into(),
                operating_status: "ONLINE".into(),
                provider: Some("octavia".into()),
            })
            .execute(&mut database_pool.get().unwrap())
            .unwrap();

        diesel::sql_query(indoc! {r#"
            CREATE TABLE pool (
                `project_id` varchar(36),
                `id` varchar(64) NOT NULL,
                `name` varchar(255),
                `protocol` varchar(16) NOT NULL,
                `lb_algorithm` varchar(255) NOT NULL,
                `operating_status` varchar(16) NOT NULL,
                `load_balancer_id` varchar(36),
                `provisioning_status` varchar(16) NOT NULL
            )"#})
        .execute(&mut database_pool.get().unwrap())
        .unwrap();
        diesel::insert_into(pool)
            .values(&Pool {
                project_id: Some("8b1632d90bfe407787d9996b7f662fd7".into()),
                id: "ca00ed86-94e3-440e-95c6-ffa35531081e".into(),
                name: Some("my_test_pool".into()),
                protocol: "TCP".into(),
                lb_algorithm: "ROUND_ROBIN".into(),
                operating_status: "ERROR".into(),
                load_balancer_id: Some("e7284bb2-f46a-42ca-8c9b-e08671255125".into()),
                provisioning_status: "ACTIVE".into(),
            })
            .execute(&mut database_pool.get().unwrap())
            .unwrap();

        diesel::sql_query(indoc! {r#"
            CREATE TABLE amphora (
                `id` varchar(36) NOT NULL,
                `compute_id` varchar(36),
                `status` varchar(36) NOT NULL,
                `load_balancer_id` varchar(36),
                `lb_network_ip` varchar(64),
                `ha_ip` varchar(64),
                `role` varchar(36),
                `cert_expiration` datetime
            )"#})
        .execute(&mut database_pool.get().unwrap())
        .unwrap();
        diesel::insert_into(amphora)
            .values(&Amphora {
                id: "45f40289-0551-483a-b089-47214bc2a8a4".into(),
                compute_id: Some("667bb225-69aa-44b1-8908-694dc624c267".into()),
                status: "READY".into(),
                load_balancer_id: Some("882f2a9d-9d53-4bd0-b0e9-08e9d0de11f9".into()),
                lb_network_ip: Some("192.168.0.6".into()),
                ha_ip: Some("10.0.0.6".into()),
                role: Some("MASTER".into()),
                cert_expiration: Some(
                    chrono::NaiveDateTime::parse_from_str(
                        "2020-08-08T23:44:31Z",
                        "%Y-%m-%dT%H:%M:%SZ",
                    )
                    .unwrap(),
                ),
            })
            .execute(&mut database_pool.get().unwrap())
            .unwrap();
        diesel::insert_into(amphora)
            .values(&Amphora {
                id: "7f890893-ced0-46ed-8697-33415d070e5a".into(),
                compute_id: Some("9cd0f9a2-fe12-42fc-a7e3-5b6fbbe20395".into()),
                status: "READY".into(),
                load_balancer_id: Some("882f2a9d-9d53-4bd0-b0e9-08e9d0de11f9".into()),
                lb_network_ip: Some("192.168.0.17".into()),
                ha_ip: Some("10.0.0.6".into()),
                role: Some("BACKUP".into()),
                cert_expiration: Some(
                    chrono::NaiveDateTime::parse_from_str(
                        "2020-08-08T23:44:30Z",
                        "%Y-%m-%dT%H:%M:%SZ",
                    )
                    .unwrap(),
                ),
            })
            .execute(&mut database_pool.get().unwrap())
            .unwrap();

        diesel::sql_query(indoc! {r#"
            CREATE TABLE vip (
                `load_balancer_id` varchar(36) NOT NULL,
                `ip_address` varchar(64)
            )"#})
        .execute(&mut database_pool.get().unwrap())
        .unwrap();
        diesel::insert_into(vip)
            .values(&VirtualIp {
                load_balancer_id: "607226db-27ef-4d41-ae89-f2a800e9c2db".into(),
                ip_address: Some("203.0.113.50".into()),
            })
            .execute(&mut database_pool.get().unwrap())
            .unwrap();

        let collector = OctaviaCollector::new(database_pool.clone());

        let mfs = collector.collect();
        assert_eq!(mfs.len(), METRICS_NUMBER);

        let encoder = TextEncoder::new();
        assert_eq!(
            encoder.encode_to_string(&mfs).unwrap(),
            indoc! {r#"
                # HELP openstack_loadbalancer_pool_status pool_status
                # TYPE openstack_loadbalancer_pool_status gauge
                openstack_loadbalancer_pool_status{id="ca00ed86-94e3-440e-95c6-ffa35531081e",lb_algorithm="ROUND_ROBIN",loadbalancers="e7284bb2-f46a-42ca-8c9b-e08671255125",name="my_test_pool",operating_status="ERROR",project_id="8b1632d90bfe407787d9996b7f662fd7",protocol="TCP",provisioning_status="ACTIVE"} 0
                # HELP openstack_loadbalancer_amphora_status amphora_status
                # TYPE openstack_loadbalancer_amphora_status gauge
                openstack_loadbalancer_amphora_status{cert_expiration="2020-08-08T23:44:30Z",compute_id="9cd0f9a2-fe12-42fc-a7e3-5b6fbbe20395",ha_ip="10.0.0.6",id="7f890893-ced0-46ed-8697-33415d070e5a",lb_network_ip="192.168.0.17",loadbalancer_id="882f2a9d-9d53-4bd0-b0e9-08e9d0de11f9",role="BACKUP",status="READY"} 2
                openstack_loadbalancer_amphora_status{cert_expiration="2020-08-08T23:44:31Z",compute_id="667bb225-69aa-44b1-8908-694dc624c267",ha_ip="10.0.0.6",id="45f40289-0551-483a-b089-47214bc2a8a4",lb_network_ip="192.168.0.6",loadbalancer_id="882f2a9d-9d53-4bd0-b0e9-08e9d0de11f9",role="MASTER",status="READY"} 2
                # HELP openstack_loadbalancer_loadbalancer_status loadbalancer_status
                # TYPE openstack_loadbalancer_loadbalancer_status gauge
                openstack_loadbalancer_loadbalancer_status{id="607226db-27ef-4d41-ae89-f2a800e9c2db",name="best_load_balancer",operating_status="ONLINE",project_id="e3cd678b11784734bc366148aa37580e",provider="octavia",provisioning_status="ACTIVE",vip_address="203.0.113.50"} 0
                # HELP openstack_loadbalancer_total_amphorae total_amphorae
                # TYPE openstack_loadbalancer_total_amphorae gauge
                openstack_loadbalancer_total_amphorae 2
                # HELP openstack_loadbalancer_total_loadbalancers total_loadbalancers
                # TYPE openstack_loadbalancer_total_loadbalancers gauge
                openstack_loadbalancer_total_loadbalancers 1
                # HELP openstack_loadbalancer_total_pools total_pools
                # TYPE openstack_loadbalancer_total_pools gauge
                openstack_loadbalancer_total_pools 1
                # HELP openstack_loadbalancer_up up
                # TYPE openstack_loadbalancer_up gauge
                openstack_loadbalancer_up 1
                "#
            }
        );
    }
}
