# Stack outputs matching ScalewayDnsZoneStackOutputs proto.

output "zone_name" {
  description = "The computed zone name (e.g., 'example.com' or 'staging.example.com')"
  value       = local.zone_name
}

output "name_servers" {
  description = "Nameservers assigned by Scaleway for this zone (configure at registrar)"
  value       = scaleway_domain_zone.zone.ns
}

output "name_servers_default" {
  description = "Scaleway's default nameservers for this zone"
  value       = scaleway_domain_zone.zone.ns_default
}

output "name_servers_master" {
  description = "Master nameservers for this zone"
  value       = scaleway_domain_zone.zone.ns_master
}

output "status" {
  description = "Zone status (e.g., 'active')"
  value       = scaleway_domain_zone.zone.status
}

# Complete outputs object matching stack_outputs.proto structure.
# Used by the Planton platform to populate status.outputs.
output "outputs" {
  description = "Complete DNS zone outputs for integration with other resources"
  value = {
    zone_name            = local.zone_name
    name_servers         = scaleway_domain_zone.zone.ns
    name_servers_default = scaleway_domain_zone.zone.ns_default
    name_servers_master  = scaleway_domain_zone.zone.ns_master
    status               = scaleway_domain_zone.zone.status
  }
}
