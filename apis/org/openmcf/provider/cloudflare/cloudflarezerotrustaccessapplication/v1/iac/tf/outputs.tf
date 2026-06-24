# outputs.tf

output "application_id" {
  description = "The unique ID of the Cloudflare Access Application"
  value       = cloudflare_zero_trust_access_application.main.id
}

output "public_hostname" {
  description = "The hostname being protected by this Access Application"
  value       = var.spec.hostname
}

output "policy_id" {
  description = "The ID of the Cloudflare Access policy associated with this application"
  value       = cloudflare_zero_trust_access_policy.main.id
}

