use diesel::r2d2::{ConnectionManager, Pool as DieselPool};
use diesel::{MultiConnection, prelude::*};
use r2d2::Error;

#[derive(MultiConnection)]
pub enum AnyConnection {
    Mysql(MysqlConnection),
    #[cfg(test)]
    Sqlite(SqliteConnection),
}

pub trait DatabaseCollector: Sized {
    fn new(pool: Pool) -> Self;

    fn connect(database_url: String) -> Result<Self, Error> {
        let pool = connect(&database_url)?;
        Ok(Self::new(pool))
    }
}

pub type Pool = DieselPool<ConnectionManager<AnyConnection>>;

pub fn connect<S: Into<String>>(database_url: S) -> Result<Pool, Error> {
    DieselPool::builder()
        .test_on_check_out(true)
        .build(ConnectionManager::<AnyConnection>::new(database_url))
}
