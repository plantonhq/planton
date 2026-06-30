# Create the DigitalOcean DNS zone (domain)
resource "digitalocean_domain" "dns_zone" {
  name = local.domain_name
}

# Create DNS records
resource "digitalocean_record" "dns_records" {
  for_each = { for record in local.dns_records : record.key => record }

  domain = digitalocean_domain.dns_zone.id
  type   = each.value.type
  name   = each.value.name
  value  = each.value.value
  ttl    = each.value.ttl_seconds

  # Priority for MX and SRV records
  priority = (
    each.value.type == "MX" || each.value.type == "SRV"
    ? coalesce(each.value.priority, 0)
    : null
  )

  # Weight for SRV records
  weight = (
    each.value.type == "SRV"
    ? coalesce(each.value.weight, 0)
    : null
  )

  # Port for SRV records
  port = (
    each.value.type == "SRV"
    ? coalesce(each.value.port, 0)
    : null
  )

  # Flags for CAA records
  flags = (
    each.value.type == "CAA"
    ? coalesce(each.value.flags, 0)
    : null
  )

  # Tag for CAA records
  tag = (
    each.value.type == "CAA"
    ? each.value.tag
    : null
  )
}

