output "zone_id" {
  description = "The Hetzner Cloud numeric ID of the created DNS zone"
  value       = hcloud_zone.this.id
}

output "nameservers" {
  description = "The authoritative Hetzner nameservers assigned to the zone"
  value       = hcloud_zone.this.authoritative_nameservers.assigned
}
