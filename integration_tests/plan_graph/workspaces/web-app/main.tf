terraform {
  required_version = ">= 1.0.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.0.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.0.0"
    }
  }
}

provider "google" {
  project = var.project
  region  = var.region
  zone    = var.zone
}

provider "random" {}

variable "project" {
  description = "The GCP project ID"
  type        = string
}

variable "region" {
  description = "The GCP region"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "The GCP zone"
  type        = string
  default     = "us-central1-a"
}

module "network" {
  source = "./modules/network"
  project = var.project
  region  = var.region
}

module "app" {
  source = "./modules/web-app"
  project = var.project
  region  = var.region
  zone    = var.zone
  network_id = module.network.network_id
  subnet_id  = module.network.subnet_id
}

# Random resources that will create changes in each plan
resource "random_id" "deployment_id" {
  byte_length = 8
  keepers = {
    # This will force the random_id to regenerate periodically
    timestamp = timestamp()
  }
}

resource "random_string" "session_token" {
  length  = 32
  special = true
  upper   = true
  lower   = true
  numeric = true
  keepers = {
    deployment_id = random_id.deployment_id.hex
  }
}

resource "random_password" "app_secret" {
  length  = 16
  special = true
  keepers = {
    deployment_id = random_id.deployment_id.hex
  }
}

resource "random_uuid" "correlation_id" {
  keepers = {
    deployment_id = random_id.deployment_id.hex
  }
}

# Local values that use random resources
locals {
  deployment_tag = "deploy-${random_id.deployment_id.hex}"
  resource_prefix = "webapp-${substr(random_id.deployment_id.hex, 0, 4)}"
  environment = "development"
  common_tags = {
    Environment = local.environment
    Deployment  = local.deployment_tag
    ManagedBy   = "terraform"
  }
}



output "network_id" {
  value = module.network.network_id
}

output "subnet_id" {
  value = module.network.subnet_id
}

output "web_instance_id" {
  value = module.app.web_instance_id
}

output "web_instance_external_ip" {
  value = module.app.web_instance_external_ip
}

output "db_instance_connection_name" {
  value = module.app.db_instance_connection_name
}

output "db_name" {
  value = module.app.db_name
}

output "db_user" {
  value = module.app.db_user
}
