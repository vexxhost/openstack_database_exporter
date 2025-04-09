mod database;
mod neutron;
mod nova;
mod nova_api;
mod octavia;

use axum::{Router, http::StatusCode, response::IntoResponse, routing::get};
use config::Config;
use database::DatabaseCollector;
use neutron::collector::NeutronCollector;
use nova::collector::NovaCollector;
use nova_api::collector::NovaApiCollector;
use octavia::collector::OctaviaCollector;
use serde::Deserialize;
use tokio::signal;

#[derive(Deserialize)]
struct ExporterConfig {
    neutron_database_url: String,
    nova_database_url: String,
    nova_api_database_url: String,
    octavia_database_url: String,
}

async fn shutdown_signal() {
    let ctrl_c = async {
        signal::ctrl_c()
            .await
            .expect("failed to install Ctrl+C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("failed to install signal handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {},
        _ = terminate => {},
    }
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let config: ExporterConfig = Config::builder()
        .add_source(config::Environment::default())
        .build()
        .unwrap()
        .try_deserialize()
        .unwrap();

    let neutron = NeutronCollector::connect(config.neutron_database_url).unwrap();
    let nova = NovaCollector::connect(config.nova_database_url).unwrap();
    let nova_api = NovaApiCollector::connect(config.nova_api_database_url).unwrap();
    let octavia = OctaviaCollector::connect(config.octavia_database_url).unwrap();

    prometheus::register(Box::new(neutron)).unwrap();
    prometheus::register(Box::new(nova)).unwrap();
    prometheus::register(Box::new(nova_api)).unwrap();
    prometheus::register(Box::new(octavia)).unwrap();

    let app = Router::new()
        .route("/", get(root))
        .route("/metrics", get(metrics));

    let listener = tokio::net::TcpListener::bind("0.0.0.0:9180").await.unwrap();
    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await
        .unwrap();
}

async fn root() -> impl IntoResponse {
    (StatusCode::OK, "OpenStack Exporter\n".to_string())
}

async fn metrics() -> impl IntoResponse {
    let encoder = prometheus::TextEncoder::new();
    let metric_families = prometheus::gather();

    (
        StatusCode::OK,
        encoder.encode_to_string(&metric_families).unwrap(),
    )
}
