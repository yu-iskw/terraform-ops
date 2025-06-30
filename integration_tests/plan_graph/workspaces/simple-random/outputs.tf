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

output "test_tag" {
  value = local.test_tag
}

output "test_prefix" {
  value = local.test_prefix
}
