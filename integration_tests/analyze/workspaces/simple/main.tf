terraform {
  required_version = ">= 1.7.0"
}

variable "secret" {
  description = "Sensitive integration-test canary"
  type        = string
  sensitive   = true
  default     = "TFOPS_INTEGRATION_SECRET_CANARY_7f3d91"
}

resource "terraform_data" "source" {
  input = {
    secret = var.secret
    name   = "source"
  }
}

resource "terraform_data" "consumer" {
  input = {
    source_id = terraform_data.source.id
  }
}

output "source_secret" {
  value     = terraform_data.source.output.secret
  sensitive = true
}
