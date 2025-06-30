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
