locals {
  # Normalize engine name to DigitalOcean API format (lowercase)
  # protobuf enum: "postgres" → "pg" (DigitalOcean uses short codes)
  # protobuf enum: "mysql" → "mysql"
  # protobuf enum: "redis" → "redis"
  # protobuf enum: "mongodb" → "mongodb"
  engine_slug_map = {
    "postgres" = "pg"
    "mysql"    = "mysql"
    "redis"    = "redis"
    "mongodb"  = "mongodb"
  }
  
  engine_slug = lookup(local.engine_slug_map, lower(var.spec.engine), lower(var.spec.engine))
  
  # Normalize region to lowercase
  region_slug = lower(var.spec.region)
  
  # Determine if VPC is specified
  has_vpc = var.spec.vpc != null && var.spec.vpc.value != ""
  
  # Extract VPC UUID if present
  vpc_uuid = local.has_vpc ? var.spec.vpc.value : null
  
  # Determine if custom storage is specified
  has_custom_storage = var.spec.storage_gib != null && var.spec.storage_gib > 0
}

