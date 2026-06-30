# Local variables for DigitalOcean Droplet module
locals {
  # Extract VPC UUID from spec
  vpc_uuid = (
    var.spec.vpc != null && var.spec.vpc.value != null
    ? var.spec.vpc.value
    : null
  )

  # Extract volume IDs from spec (filter out nulls)
  volume_ids = compact([
    for vol in coalesce(var.spec.volume_ids, []) : (
      vol.value != null ? vol.value : null
    )
  ])

  # Map protobuf region enum to DigitalOcean region slug
  region_map = {
    "digital_ocean_region_unspecified" = "nyc3"  # Default fallback
    "digital_ocean_region_nyc1"        = "nyc1"
    "digital_ocean_region_nyc2"        = "nyc2"
    "digital_ocean_region_nyc3"        = "nyc3"
    "digital_ocean_region_sfo1"        = "sfo1"
    "digital_ocean_region_sfo2"        = "sfo2"
    "digital_ocean_region_sfo3"        = "sfo3"
    "digital_ocean_region_ams2"        = "ams2"
    "digital_ocean_region_ams3"        = "ams3"
    "digital_ocean_region_sgp1"        = "sgp1"
    "digital_ocean_region_lon1"        = "lon1"
    "digital_ocean_region_fra1"        = "fra1"
    "digital_ocean_region_tor1"        = "tor1"
    "digital_ocean_region_blr1"        = "blr1"
    "digital_ocean_region_syd1"        = "syd1"
  }

  # Resolve region slug
  region_slug = local.region_map[var.spec.region]

  # Combine user-provided tags with metadata tags
  tags = distinct(concat(
    coalesce(var.spec.tags, []),
    [
      "managed-by:planton",
      "resource-kind:digitalocean-droplet",
      "resource-name:${var.metadata.name}"
    ]
  ))

  # Monitoring flag (inverted from disable_monitoring)
  monitoring = !coalesce(var.spec.disable_monitoring, false)

  # Backups flag
  backups = coalesce(var.spec.enable_backups, false)

  # IPv6 flag
  ipv6 = coalesce(var.spec.enable_ipv6, false)
}

