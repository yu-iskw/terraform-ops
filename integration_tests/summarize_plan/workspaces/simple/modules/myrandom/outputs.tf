output "test_integer" {
  value     = random_integer.test_integer.result
  sensitive = true
}

output "test_string" {
  value     = random_string.test_string.result
  sensitive = false
}
