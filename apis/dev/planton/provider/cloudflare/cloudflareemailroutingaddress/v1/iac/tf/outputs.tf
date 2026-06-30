output "address_id" {
  description = "The Cloudflare-assigned identifier of the destination address"
  value       = cloudflare_email_routing_address.main.id
}

output "email" {
  description = "The destination email address"
  value       = cloudflare_email_routing_address.main.email
}

output "verified" {
  description = "RFC3339 timestamp of verification, or empty if not yet verified"
  value       = cloudflare_email_routing_address.main.verified
}

output "created" {
  description = "RFC3339 timestamp of when the address was created"
  value       = cloudflare_email_routing_address.main.created
}
