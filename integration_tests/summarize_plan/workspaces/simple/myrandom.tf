module "myrandom" {
  source = "./modules/myrandom"
}

output "myrandom" {
  description = "My random module"
  value       = module.myrandom
  sensitive   = true
}
