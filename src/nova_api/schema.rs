use diesel::table;

table! {
  build_requests (id) {
      id -> Integer,
      #[max_length = 255]
      project_id -> Varchar,
      #[max_length = 36]
      instance_uuid -> Nullable<Varchar>,
  }
}
