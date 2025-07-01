resource "random_integer" "test_integer" {
  min = 1
  max = 100
}

resource "random_string" "test_string" {
  length = 10
  special = false
  upper = false
}
