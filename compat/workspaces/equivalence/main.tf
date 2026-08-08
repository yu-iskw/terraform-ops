terraform {
  required_version = ">= 1.7.0"
}

variable "secret" {
  type      = string
  default   = "compatibility-secret-canary"
  sensitive = true
}

resource "terraform_data" "source" {
  input = {
    secret = var.secret
  }
}

module "child" {
  source = "./modules/child"
  value  = terraform_data.source.output
}

resource "terraform_data" "consumer" {
  input = module.child.result
}

output "result" {
  value     = terraform_data.consumer.output
  sensitive = true
}
