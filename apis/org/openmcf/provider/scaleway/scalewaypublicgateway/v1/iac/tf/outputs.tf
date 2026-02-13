# Gateway ID
output "gateway_id" {
  description = "The unique identifier of the created Public Gateway"
  value       = scaleway_vpc_public_gateway.gateway.id
}

# Public IP Address
output "public_ip_address" {
  description = "The public IPv4 address assigned to the gateway"
  value       = scaleway_vpc_public_gateway_ip.ip.address
}

# Public IP ID
output "public_ip_id" {
  description = "The unique identifier of the Flexible IP resource"
  value       = scaleway_vpc_public_gateway_ip.ip.id
}

# Gateway Network ID
output "gateway_network_id" {
  description = "The unique identifier of the gateway-to-network attachment"
  value       = scaleway_vpc_gateway_network.attachment.id
}

# Gateway Status
output "gateway_status" {
  description = "The operational status of the Public Gateway"
  value       = scaleway_vpc_public_gateway.gateway.status
}

# Organization ID
output "organization_id" {
  description = "The Organization ID the gateway is associated with"
  value       = scaleway_vpc_public_gateway.gateway.organization_id
}

# Zone
output "zone" {
  description = "The zone where the gateway is deployed"
  value       = scaleway_vpc_public_gateway.gateway.zone
}

# Gateway Name
output "gateway_name" {
  description = "The name of the Public Gateway"
  value       = scaleway_vpc_public_gateway.gateway.name
}
