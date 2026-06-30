output "rule_id" {
  description = "The Cloudflare-assigned identifier of the routing rule"
  value       = cloudflare_email_routing_rule.main.id
}

output "zone_id" {
  description = "The zone ID the rule belongs to"
  value       = cloudflare_email_routing_rule.main.zone_id
}
