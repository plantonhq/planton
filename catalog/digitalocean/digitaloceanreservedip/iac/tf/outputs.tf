# Outputs — exactly the DigitalOceanReservedIpOutputs contract,
# identical across both provisioners. The address IS the resource identity
# (imports and API lookups address the reservation by it).

output "reserved_ip_address" {
  description = "The reserved IP address (IPv4 or IPv6 per the spec's ip_version)"
  value       = local.is_ipv6 ? digitalocean_reserved_ipv6.ipv6[0].ip : digitalocean_reserved_ip.ipv4[0].ip_address
}

output "urn" {
  description = "The uniform resource name of the reservation, as used by DigitalOcean project membership"
  value       = local.is_ipv6 ? digitalocean_reserved_ipv6.ipv6[0].urn : digitalocean_reserved_ip.ipv4[0].urn
}
