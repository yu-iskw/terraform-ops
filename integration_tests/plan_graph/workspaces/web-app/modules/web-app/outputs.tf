output "web_instance_id" {
  value = google_compute_instance.web.id
}

output "web_instance_external_ip" {
  value = google_compute_instance.web.network_interface[0].access_config[0].nat_ip
}

output "db_instance_connection_name" {
  value = module.database.db_instance_connection_name
}

output "db_name" {
  value = module.database.db_name
}

output "db_user" {
  value = module.database.db_user
}
