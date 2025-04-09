use diesel::table;

table! {
  instances (id) {
      id -> Integer,
      #[max_length = 36]
      uuid -> Varchar,
      #[max_length = 255]
      task_state -> Nullable<Varchar>,
      deleted -> Nullable<Integer>,
  }
}
