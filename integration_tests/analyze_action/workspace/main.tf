terraform {
  required_version = ">= 1.7.0"
}

variable "secret" {
  type      = string
  default   = "TFOPS_ACTION_SECRET_CANARY_4ddbe6"
  sensitive = true
}

resource "terraform_data" "example" {
  input = {
    secret = var.secret
  }
}
