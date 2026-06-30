# Create the DigitalOcean DNS record
resource "digitalocean_record" "dns_record" {
  domain = local.domain
  type   = local.type
  name   = local.name
  value  = local.value
  ttl    = local.ttl_seconds

  # Priority for MX and SRV records
  priority = (
    local.type == "MX" || local.type == "SRV"
    ? coalesce(local.priority, 0)
    : null
  )

  # Weight for SRV records
  weight = (
    local.type == "SRV"
    ? coalesce(local.weight, 0)
    : null
  )

  # Port for SRV records
  port = (
    local.type == "SRV"
    ? coalesce(local.port, 0)
    : null
  )

  # Flags for CAA records
  flags = (
    local.type == "CAA"
    ? coalesce(local.flags, 0)
    : null
  )

  # Tag for CAA records
  tag = (
    local.type == "CAA"
    ? local.tag
    : null
  )
}
