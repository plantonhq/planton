# VPC ID
output "vpc_id" {
  description = "The unique identifier (UUID) of the created Scaleway VPC"
  value       = scaleway_vpc.vpc.id
}

# Is Default
output "is_default" {
  description = "Whether this VPC is the default one for its project"
  value       = scaleway_vpc.vpc.is_default
}

# Organization ID
output "organization_id" {
  description = "The Organization ID the VPC is associated with"
  value       = scaleway_vpc.vpc.organization_id
}

# Created At
output "created_at" {
  description = "Timestamp when the VPC was created (RFC 3339)"
  value       = scaleway_vpc.vpc.created_at
}

# Region
output "region" {
  description = "The region where the VPC is deployed"
  value       = scaleway_vpc.vpc.region
}

# VPC Name
output "vpc_name" {
  description = "The name of the VPC"
  value       = scaleway_vpc.vpc.name
}
