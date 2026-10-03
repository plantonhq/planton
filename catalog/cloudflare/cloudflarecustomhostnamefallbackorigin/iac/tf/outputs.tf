# No status output: fallback-origin deployment is asynchronous
# (pending_deployment -> active), and a point-in-time phase is never a stable
# output -- it flips on the first refresh after the transition and
# re-plans forever. Read deployment status from the Cloudflare API instead.

output "created_at" {
  description = "RFC3339 timestamp of when the fallback origin was created"
  value       = cloudflare_custom_hostname_fallback_origin.main.created_at
}

output "updated_at" {
  description = "RFC3339 timestamp of when the fallback origin was last updated"
  value       = cloudflare_custom_hostname_fallback_origin.main.updated_at
}

# No errors output: like the sibling custom hostname's verification_errors
# (measured live 2026-08-29), the server populates and clears this list
# asynchronously after apply -- a transient diagnostic is never a stable
# output. Read deployment errors from the Cloudflare API instead.

output "zone_id" {
  description = "The Cloudflare zone this singleton belongs to (the fallback origin has no resource id; its API identity IS the zone)"
  value       = cloudflare_custom_hostname_fallback_origin.main.zone_id
}
