# VPC ID
output "vpc_id" {
  description = "The unique identifier (UUID) of the created VPC"
  value       = digitalocean_vpc.vpc.id
}

# VPC URN
output "vpc_urn" {
  description = "The uniform resource name (URN) of the VPC"
  value       = digitalocean_vpc.vpc.urn
}

# IP Range (actual)
output "ip_range" {
  description = "The actual IP range assigned to the VPC (auto-generated or explicitly specified)"
  value       = digitalocean_vpc.vpc.ip_range
}

# Is Default
output "is_default" {
  description = "Whether this VPC is the default for its region"
  value       = digitalocean_vpc.vpc.default
}

# Created At
output "created_at" {
  description = "Timestamp when the VPC was created"
  value       = digitalocean_vpc.vpc.created_at
}

# Region
output "region" {
  description = "The region where the VPC is deployed"
  value       = digitalocean_vpc.vpc.region
}

# VPC Name
output "vpc_name" {
  description = "The name of the VPC"
  value       = digitalocean_vpc.vpc.name
}

