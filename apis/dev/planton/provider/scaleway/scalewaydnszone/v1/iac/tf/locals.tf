locals {
  # ── Zone identity ──────────────────────────────────────────────────
  domain    = var.spec.domain
  subdomain = var.spec.subdomain

  # Computed zone name: "{subdomain}.{domain}" or just "{domain}" for root.
  zone_name = local.subdomain != "" ? "${local.subdomain}.${local.domain}" : local.domain

  # ── Record type mapping ────────────────────────────────────────────
  # Maps local RecordType proto enum values to Scaleway API record type
  # strings. Includes all 13 Scaleway-supported record types.
  record_type_map = {
    "record_type_unspecified" = "A" # Fallback, should not occur with proper validation
    "A"                       = "A"
    "AAAA"                    = "AAAA"
    "ALIAS"                   = "ALIAS"
    "CAA"                     = "CAA"
    "CNAME"                   = "CNAME"
    "DNAME"                   = "DNAME"
    "MX"                      = "MX"
    "NS"                      = "NS"
    "PTR"                     = "PTR"
    "SOA"                     = "SOA"
    "SRV"                     = "SRV"
    "TXT"                     = "TXT"
    "TLSA"                    = "TLSA"
  }

  # ── Record flattening ──────────────────────────────────────────────
  # Flatten inline records into a map suitable for for_each.
  # Each record entry creates one scaleway_domain_record resource.
  dns_records = {
    for idx, record in coalesce(var.spec.records, []) :
    "${coalesce(record.name, "apex")}-${idx}" => {
      name     = coalesce(record.name, "")
      type     = lookup(local.record_type_map, record.type, record.type)
      data     = record.data.value
      ttl      = coalesce(record.ttl, 3600)
      priority = coalesce(record.priority, 0)
    }
  }

  # NOTE: Scaleway DNS zones and records do not support tags.
  # Unlike most other Scaleway resources, the DNS API does not accept
  # tags/labels. Standard Planton metadata tags are not applied.
}
