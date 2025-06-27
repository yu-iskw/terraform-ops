# Random resource outputs that will change between plan runs
output "deployment_id" {
  description = "Random deployment identifier"
  value       = random_id.deployment_id.hex
}

output "session_token" {
  description = "Random session token"
  value       = random_string.session_token.result
  sensitive   = true
}

output "app_secret" {
  description = "Random application secret"
  value       = random_password.app_secret.result
  sensitive   = true
}

output "correlation_id" {
  description = "Random correlation UUID"
  value       = random_uuid.correlation_id.result
}

output "deployment_tag" {
  description = "Deployment tag using random ID"
  value       = local.deployment_tag
}

output "resource_prefix" {
  description = "Resource prefix using random ID"
  value       = local.resource_prefix
}

# Summary output showing all random values
output "random_summary" {
  description = "Summary of all random values for this deployment"
  value = {
    deployment_id   = random_id.deployment_id.hex
    deployment_tag  = local.deployment_tag
    resource_prefix = local.resource_prefix
    correlation_id  = random_uuid.correlation_id.result
    # Sensitive values excluded from summary
  }
}
