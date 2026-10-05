output "custom_hostname_id" {
  description = "The custom hostname identifier"
  value       = cloudflare_custom_hostname.main.id
}

# No status output: hostname activation is asynchronous (pending ->
# pending_validation -> active), and a point-in-time phase is never a stable
# output -- it flips on the first refresh after the transition and
# re-plans forever. Read activation status from the Cloudflare API instead.

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

# No verification_errors output: the server populates the list
# asynchronously after apply ("zone is not active yet" measured appearing
# seconds post-create) and clears it on activation -- a transient diagnostic
# is never a stable output (output-only changes fail idempotent
# re-plans). Read it from the Cloudflare API instead.

output "created_at" {
  description = "RFC3339 timestamp of when the custom hostname was created"
  value       = cloudflare_custom_hostname.main.created_at
}

output "zone_id" {
  description = "The Cloudflare zone the hostname was onboarded onto (a custom hostname's API identity is zone_id + custom_hostname_id)"
  value       = cloudflare_custom_hostname.main.zone_id
}
