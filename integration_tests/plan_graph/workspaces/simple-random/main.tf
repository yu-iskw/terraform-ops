terraform {
  required_version = ">= 1.0.0"
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = ">= 3.0.0"
    }
  }
}

provider "random" {}

# Simple random resources for testing
resource "random_id" "test_id" {
  byte_length = 4
  keepers = {
    timestamp = timestamp()
  }
}

resource "random_string" "test_string" {
  length  = 16
  special = false
  upper   = true
  lower   = true
  numeric = true
  keepers = {
    test_id = random_id.test_id.hex
  }
}

resource "random_password" "test_password" {
  length  = 12
  special = true
  keepers = {
    test_id = random_id.test_id.hex
  }
}

resource "random_uuid" "test_uuid" {
  keepers = {
    test_id = random_id.test_id.hex
  }
}

resource "random_integer" "test_integer" {
  min = 1
  max = 100
  keepers = {
    test_id = random_id.test_id.hex
  }
}

resource "random_pet" "test_pet" {
  keepers = {
    test_id = random_id.test_id.hex
  }
}

# Local values
locals {
  test_tag = "test-${random_id.test_id.hex}"
  test_prefix = "simple-${substr(random_id.test_id.hex, 0, 2)}"
}

# Outputs
output "test_id" {
  value = random_id.test_id.hex
}

output "test_string" {
  value = random_string.test_string.result
}

output "test_password" {
  value     = random_password.test_password.result
  sensitive = true
}

output "test_uuid" {
  value = random_uuid.test_uuid.result
}

output "test_integer" {
  value = random_integer.test_integer.result
}

output "test_pet" {
  value = random_pet.test_pet.id
}

output "test_tag" {
  value = local.test_tag
}

output "test_prefix" {
  value = local.test_prefix
}
