# Outputs, exactly the DigitalOceanDropletOutputs contract.
# Live state (status, locked) is deliberately not exported: apply-time
# snapshots go stale, and verification reads the live API.

output "droplet_id" {
  description = "The unique identifier of the created Droplet"
  value       = digitalocean_droplet.this.id
}

output "ipv4_address" {
  description = "The public IPv4 address of the Droplet"
  value       = digitalocean_droplet.this.ipv4_address
}

output "ipv6_address" {
  description = "The public IPv6 address (empty unless IPv6 is enabled)"
  value       = digitalocean_droplet.this.ipv6_address
}

output "ipv4_address_private" {
  description = "The private IPv4 address inside the Droplet's VPC"
  value       = digitalocean_droplet.this.ipv4_address_private
}

output "urn" {
  description = "The uniform resource name, accepted by projects and firewalls as a member reference"
  value       = digitalocean_droplet.this.urn
}

output "vpc_uuid" {
  description = "The UUID of the VPC the Droplet landed in (the region's default VPC when spec.vpc was omitted)"
  value       = digitalocean_droplet.this.vpc_uuid
}
