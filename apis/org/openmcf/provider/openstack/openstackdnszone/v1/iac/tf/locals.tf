# locals.tf

locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # The DNS zone name is derived from spec.domain_name
  zone_name = var.spec.domain_name

  # Build a map of inline records keyed by "type-record_name" for stable for_each.
  # This provides stable IaC state -- adding/removing/reordering records only
  # affects the specific record being changed, not others.
  records_map = {
    for r in var.spec.records :
    "${r.record_type}-${r.record_name}" => r
  }

  # Map of record type numbers to their string names
  # (matches the OpenStackDnsRecord.RecordType proto enum values).
  record_type_names = {
    1  = "A"
    2  = "AAAA"
    3  = "CNAME"
    4  = "MX"
    5  = "TXT"
    6  = "SRV"
    7  = "NS"
    8  = "PTR"
    9  = "CAA"
    10 = "SOA"
    11 = "SPF"
    12 = "SSHFP"
    13 = "NAPTR"
  }
}
