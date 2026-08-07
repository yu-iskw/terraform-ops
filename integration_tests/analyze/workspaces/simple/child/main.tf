variable "source_id" {
  type = string
}

resource "terraform_data" "child" {
  input = {
    source_id = var.source_id
  }
}

output "result" {
  value = terraform_data.child.id
}
