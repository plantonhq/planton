# DigitalOcean VPC Resource
resource "digitalocean_vpc" "vpc" {
  name   = local.vpc_name
  region = local.region
  
  # Optional description
  description = local.description != "" ? local.description : null
  
  # Optional IP range (80/20 principle)
  # When omitted (empty string), DigitalOcean auto-generates a non-conflicting /20 CIDR
  # When specified, uses the explicit CIDR block for production IP planning
  ip_range = local.use_explicit_cidr ? local.ip_range_cidr : null
  
  # Lifecycle management
  lifecycle {
    # VPC IP ranges are immutable - prevents accidental changes
    # If you need to change IP range, must create new VPC
    ignore_changes = [
      ip_range
    ]
  }
}

