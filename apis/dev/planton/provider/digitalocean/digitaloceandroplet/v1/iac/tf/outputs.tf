# Output the Droplet ID
output "droplet_id" {
  description = "The unique identifier of the created Droplet"
  value       = digitalocean_droplet.droplet.id
}

# Output the IPv4 address
output "ipv4_address" {
  description = "The public IPv4 address of the Droplet"
  value       = digitalocean_droplet.droplet.ipv4_address
}

# Output the IPv6 address (if enabled)
output "ipv6_address" {
  description = "The public IPv6 address of the Droplet (null if IPv6 not enabled)"
  value       = digitalocean_droplet.droplet.ipv6_address
}

# Output the private IPv4 address (VPC)
output "ipv4_address_private" {
  description = "The private IPv4 address within the VPC"
  value       = digitalocean_droplet.droplet.ipv4_address_private
}

# Output the image ID
output "image_id" {
  description = "The image ID or slug used to create the Droplet"
  value       = digitalocean_droplet.droplet.image
}

# Output the VPC UUID
output "vpc_uuid" {
  description = "The VPC UUID the Droplet is assigned to"
  value       = digitalocean_droplet.droplet.vpc_uuid
}

# Output the Droplet URN
output "urn" {
  description = "The uniform resource name (URN) of the Droplet"
  value       = digitalocean_droplet.droplet.urn
}

# Output status
output "status" {
  description = "The status of the Droplet (active, off, archive)"
  value       = digitalocean_droplet.droplet.status
}

# Output all tags
output "tags" {
  description = "The tags applied to the Droplet"
  value       = digitalocean_droplet.droplet.tags
}

