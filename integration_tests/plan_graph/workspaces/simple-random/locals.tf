# Local values
locals {
  test_tag    = "test-${random_id.test_id.hex}"
  test_prefix = "simple-${substr(random_id.test_id.hex, 0, 2)}"
}
