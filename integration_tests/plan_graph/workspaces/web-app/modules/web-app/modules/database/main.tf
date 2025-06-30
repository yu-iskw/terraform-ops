variable "project" { type = string }
variable "region" { type = string }

resource "google_sql_database_instance" "main" {
  name             = "web-app-db"
  database_version = "POSTGRES_14"
  region           = var.region
  project          = var.project

  settings {
    tier = "db-f1-micro"
    ip_configuration {
      authorized_networks {
        value = "0.0.0.0/0"
      }
      ipv4_enabled = true
    }
    backup_configuration {
      enabled = true
    }
  }
}

resource "google_sql_database" "app" {
  name     = "webappdb"
  instance = google_sql_database_instance.main.name
  project  = var.project
}

resource "google_sql_user" "app" {
  name     = "dbadmin"
  instance = google_sql_database_instance.main.name
  password = "temporary-password-123"
  project  = var.project
}
