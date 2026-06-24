# outputs.tf

output "zone_id" {
  description = "The unique identifier of the created Cloudflare zone"
  value       = cloudflare_zone.main.id
}

output "nameservers" {
  description = "The Cloudflare nameservers assigned to this zone"
  value       = cloudflare_zone.main.name_servers
}
