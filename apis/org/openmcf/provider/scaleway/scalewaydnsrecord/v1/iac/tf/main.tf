# ScalewayDnsRecord Terraform Module
#
# Standalone resource: creates a single Scaleway DNS record in an
# existing DNS zone.
#
# Resources created:
#   - scaleway_domain_record (1x) -- the DNS record
#
# NOTE: Scaleway DNS records do not support tags. Unlike most other
# Scaleway resources, the DNS API does not accept tags/labels.

resource "scaleway_domain_record" "record" {
  dns_zone = local.zone_name
  name     = local.name
  type     = local.type
  data     = local.data
  ttl      = local.ttl

  # Priority is only meaningful for MX and SRV records.
  # Set to null for other types to avoid unnecessary API calls.
  priority = (
    local.type == "MX" || local.type == "SRV"
    ? local.priority
    : null
  )

  # NOTE: keep_empty_zone is documented in the Scaleway TF provider docs
  # but is not supported in the current provider version (v2.69.0). The
  # spec field is preserved for forward compatibility -- when a future
  # provider version exposes this argument, uncomment the line below:
  # keep_empty_zone = local.keep_empty_zone
}
