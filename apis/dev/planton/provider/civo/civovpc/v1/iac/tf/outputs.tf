output "network_id" {
  description = "The unique identifier (ID) of the created Civo network"
  value       = civo_network.main.id
}

output "cidr_block" {
  description = "The IPv4 CIDR block of the created network"
  value       = civo_network.main.cidr_v4
}

output "created_at_rfc3339" {
  description = "Timestamp when the network was created (RFC 3339 format)"
  value       = ""
  # Note: The Civo provider doesn't expose the created_at timestamp as an attribute.
  # This output is defined for API consistency but will be empty.
  # The network exists after creation, but the timestamp is not available via the provider.
}

output "network_name" {
  description = "The name (label) of the created network"
  value       = civo_network.main.label
}

output "region" {
  description = "The region where the network was created"
  value       = local.region
}

output "is_default" {
  description = "Whether this network is the default network for the region"
  value       = civo_network.main.default
}

