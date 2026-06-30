# Private Network ID
output "private_network_id" {
  description = "The unique identifier (UUID) of the created Scaleway Private Network"
  value       = scaleway_vpc_private_network.private_network.id
}

# IPv4 Subnet CIDR
output "ipv4_subnet_cidr" {
  description = "The IPv4 subnet CIDR associated with the Private Network (specified or auto-allocated)"
  value       = try(scaleway_vpc_private_network.private_network.ipv4_subnet[0].subnet, "")
}

# Organization ID
output "organization_id" {
  description = "The Organization ID the Private Network is associated with"
  value       = scaleway_vpc_private_network.private_network.organization_id
}

# Created At
output "created_at" {
  description = "Timestamp when the Private Network was created (RFC 3339)"
  value       = scaleway_vpc_private_network.private_network.created_at
}

# Region
output "region" {
  description = "The region where the Private Network is deployed"
  value       = scaleway_vpc_private_network.private_network.region
}

# Private Network Name
output "private_network_name" {
  description = "The name of the Private Network"
  value       = scaleway_vpc_private_network.private_network.name
}
