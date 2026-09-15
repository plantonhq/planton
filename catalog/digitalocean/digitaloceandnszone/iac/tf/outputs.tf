output "zone_name" {
  description = "The domain name of the DNS zone"
  value       = digitalocean_domain.dns_zone.name
}

output "zone_id" {
  description = "The zone's resource identifier — DigitalOcean addresses domains by name, so this is the domain name itself"
  value       = digitalocean_domain.dns_zone.id
}

output "name_servers" {
  description = "DigitalOcean's authoritative name servers (a fixed platform-wide set the API does not return per zone); set these at the registrar to delegate"
  value = [
    "ns1.digitalocean.com",
    "ns2.digitalocean.com",
    "ns3.digitalocean.com"
  ]
}

output "urn" {
  description = "The uniform resource name of the domain (e.g. do:domain:example.com)"
  value       = digitalocean_domain.dns_zone.urn
}

output "record_ids" {
  description = "Numeric ids of the inline records, keyed by the module's for_each key (<record name>-<record index>-<value index>) -- the handles the API addresses each record by and state import takes as the second half of {domain},{record_id}"
  value       = { for k, record in digitalocean_record.dns_records : k => record.id }
}
