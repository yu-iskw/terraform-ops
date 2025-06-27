output "db_instance_connection_name" {
  value = google_sql_database_instance.main.connection_name
}

output "db_name" {
  value = google_sql_database.app.name
}

output "db_user" {
  value = google_sql_user.app.name
}
