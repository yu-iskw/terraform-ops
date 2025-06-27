variable "project" { type = string }
variable "region" { type = string }

resource "google_compute_network" "main" {
  name                    = "web-app-vpc"
  auto_create_subnetworks = false
  project                 = var.project
}

resource "google_compute_subnetwork" "public" {
  name          = "web-app-public-subnet"
  ip_cidr_range = "10.0.1.0/24"
  region        = var.region
  network       = google_compute_network.main.id
  project       = var.project
}

output "network_id" {
  value = google_compute_network.main.id
}

output "subnet_id" {
  value = google_compute_subnetwork.public.id
}
