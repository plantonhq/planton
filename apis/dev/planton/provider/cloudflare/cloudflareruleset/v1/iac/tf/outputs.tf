output "ruleset_id" {
  description = "The Cloudflare-assigned unique identifier of the ruleset"
  value       = cloudflare_ruleset.main.id
}

output "version" {
  description = "The current version of the ruleset"
  value       = cloudflare_ruleset.main.version
}

output "zone_id" {
  description = "The zone ID the ruleset belongs to (pass-through)"
  value       = local.zone_id
}

output "phase" {
  description = "The phase the ruleset executes in (pass-through)"
  value       = local.phase
}

output "last_updated" {
  description = "RFC3339 timestamp of the ruleset's last update"
  value       = cloudflare_ruleset.main.last_updated
}
