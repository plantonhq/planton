output "zone_id" {
  description = "The zone ID Email Routing was enabled on"
  value       = cloudflare_email_routing_settings.main.zone_id
}

output "enabled" {
  description = "Whether Email Routing is enabled on the zone"
  value       = cloudflare_email_routing_settings.main.enabled
}

output "status" {
  description = "The Email Routing configuration status"
  value       = cloudflare_email_routing_settings.main.status
}

output "name" {
  description = "The zone's domain name"
  value       = cloudflare_email_routing_settings.main.name
}
