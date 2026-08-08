variable "value" {
  type      = any
  sensitive = true
}

resource "terraform_data" "inner" {
  input = var.value
}

output "result" {
  value     = terraform_data.inner.output
  sensitive = true
}
