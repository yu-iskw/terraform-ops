variable "project" { type = string }
variable "region" { type = string }
variable "zone" { type = string }
variable "network_id" { type = string }
variable "subnet_id" { type = string }

resource "google_compute_firewall" "web" {
  name    = "web-allow-http-https-ssh"
  network = var.network_id

  allow {
    protocol = "tcp"
    ports    = ["80", "443", "22"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["web-server"]
  project       = var.project
}

resource "google_compute_instance" "web" {
  name         = "web-app-instance"
  machine_type = "e2-micro"
  zone         = var.zone
  project      = var.project

  tags = ["web-server"]

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
    }
  }

  network_interface {
    network    = var.network_id
    subnetwork = var.subnet_id
    access_config {}
  }

  metadata_startup_script = <<-EOT
    #!/bin/bash
    apt-get update -y
    apt-get install -y apache2
    systemctl start apache2
    echo "<h1>Hello from Terraform on GCP!</h1>" > /var/www/html/index.html
  EOT
}

module "database" {
  source  = "./modules/database"
  project = var.project
  region  = var.region
}
