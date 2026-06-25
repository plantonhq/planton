output "custom_hostname_id" {
  description = "The custom hostname identifier"
  value       = cloudflare_custom_hostname.main.id
}

output "status" {
  description = "The activation status"
  value       = cloudflare_custom_hostname.main.status
}

output "ownership_verification_name" {
  description = "The DNS record name for ownership verification"
  value       = try(cloudflare_custom_hostname.main.ownership_verification.name, "")
}

output "ownership_verification_type" {
  description = "The DNS record type for ownership verification"
  value       = try(cloudflare_custom_hostname.main.ownership_verification.type, "")
}

output "ownership_verification_value" {
  description = "The DNS record value for ownership verification"
  value       = try(cloudflare_custom_hostname.main.ownership_verification.value, "")
}

output "ownership_verification_http_url" {
  description = "The HTTP verification URL"
  value       = try(cloudflare_custom_hostname.main.ownership_verification_http.http_url, "")
}

output "ownership_verification_http_body" {
  description = "The body served at the HTTP verification URL"
  value       = try(cloudflare_custom_hostname.main.ownership_verification_http.http_body, "")
}

output "verification_errors" {
  description = "Any verification errors reported by Cloudflare"
  value       = try(cloudflare_custom_hostname.main.verification_errors, [])
}

output "created_at" {
  description = "RFC3339 timestamp of when the custom hostname was created"
  value       = cloudflare_custom_hostname.main.created_at
}
