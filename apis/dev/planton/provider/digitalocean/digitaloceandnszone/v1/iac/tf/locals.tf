# Local variables for DigitalOcean DNS Zone module
locals {
  # Extract domain name from spec
  domain_name = var.spec.domain_name

  # Process DNS records to flatten for iteration
  # Each record with multiple values becomes multiple DigitalOcean DNS records
  dns_records = flatten([
    for idx, record in coalesce(var.spec.records, []) : [
      for val_idx, value in record.values : {
        # Unique identifier for this specific record instance
        key = "${record.name}-${idx}-${val_idx}"

        # Record configuration
        name        = record.name
        type        = local.record_type_map[record.type]
        value       = value.value
        ttl_seconds = coalesce(record.ttl_seconds, 3600)

        # MX and SRV priority
        priority = record.priority

        # SRV-specific fields
        weight = record.weight
        port   = record.port

        # CAA-specific fields
        flags = record.flags
        tag   = record.tag
      }
    ]
  ])

  # Map protobuf enum values to Terraform/DigitalOcean record types
  record_type_map = {
    "dns_record_type_unspecified" = "A"           # Default fallback
    "dns_record_type_a"           = "A"
    "dns_record_type_aaaa"        = "AAAA"
    "dns_record_type_cname"       = "CNAME"
    "dns_record_type_mx"          = "MX"
    "dns_record_type_txt"         = "TXT"
    "dns_record_type_srv"         = "SRV"
    "dns_record_type_caa"         = "CAA"
    "dns_record_type_ns"          = "NS"
    "dns_record_type_soa"         = "SOA"
    "dns_record_type_ptr"         = "PTR"
  }

  # Metadata labels for tagging/tracking
  labels = merge(
    {
      "managed-by"    = "planton"
      "resource-kind" = "digitalocean-dns-zone"
      "resource-name" = var.metadata.name
    },
    coalesce(var.metadata.labels, {})
  )
}

