# Output the zone name
output "zone_name" {
  description = "The domain name of the created DNS zone"
  value       = digitalocean_domain.dns_zone.name
}

# Output the zone ID (same as name for DigitalOcean)
output "zone_id" {
  description = "The ID of the created DNS zone"
  value       = digitalocean_domain.dns_zone.id
}

# Output the DigitalOcean nameservers
output "name_servers" {
  description = "DigitalOcean nameservers for this zone"
  value = [
    "ns1.digitalocean.com",
    "ns2.digitalocean.com",
    "ns3.digitalocean.com"
  ]
}

# Output the URN (DigitalOcean Universal Resource Name)
output "urn" {
  description = "The uniform resource name (URN) of the domain"
  value       = digitalocean_domain.dns_zone.urn
}

# Output all created DNS records for reference
output "dns_records" {
  description = "Map of all created DNS records"
  value = {
    for k, v in digitalocean_record.dns_records : k => {
      id    = v.id
      fqdn  = v.fqdn
      type  = v.type
      name  = v.name
      value = v.value
    }
  }
}

