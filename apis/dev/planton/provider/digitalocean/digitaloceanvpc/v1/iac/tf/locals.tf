locals {
  # VPC name from metadata
  vpc_name = var.metadata.name
  
  # Region from spec
  region = var.spec.region
  
  # Description (optional)
  description = var.spec.description
  
  # IP range CIDR (optional - 80/20 principle)
  # When empty, DigitalOcean auto-generates a non-conflicting /20 block
  ip_range_cidr = var.spec.ip_range_cidr
  
  # Whether to use explicit IP range (20% use case) vs auto-generate (80% use case)
  use_explicit_cidr = var.spec.ip_range_cidr != ""
}

