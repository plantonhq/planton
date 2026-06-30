locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base tags for Azure resources
  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "azure_dns_record"
    "resource_name" = var.metadata.name
  }

  # Organization tag only if var.metadata.org is non-empty
  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  # Environment tag only if var.metadata.env is non-empty
  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  # Merge base, org, and environment tags
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag)

  # Record type checks
  is_a_record     = var.spec.type == "A"
  is_aaaa_record  = var.spec.type == "AAAA"
  is_cname_record = var.spec.type == "CNAME"
  is_mx_record    = var.spec.type == "MX"
  is_txt_record   = var.spec.type == "TXT"
  is_ns_record    = var.spec.type == "NS"
  is_caa_record   = var.spec.type == "CAA"
  is_srv_record   = var.spec.type == "SRV"
  is_ptr_record   = var.spec.type == "PTR"

  # Compute FQDN for the record
  fqdn = var.spec.name == "@" ? var.spec.zone_name : "${var.spec.name}.${var.spec.zone_name}"
}
