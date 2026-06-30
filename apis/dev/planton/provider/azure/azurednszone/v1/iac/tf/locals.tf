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
    "resource_kind" = "azure_dns_zone"
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

  # Group records by type for easier management
  a_records     = [for r in var.spec.records : r if r.record_type == "A"]
  aaaa_records  = [for r in var.spec.records : r if r.record_type == "AAAA"]
  cname_records = [for r in var.spec.records : r if r.record_type == "CNAME"]
  mx_records    = [for r in var.spec.records : r if r.record_type == "MX"]
  txt_records   = [for r in var.spec.records : r if r.record_type == "TXT"]
  ns_records    = [for r in var.spec.records : r if r.record_type == "NS"]
  caa_records   = [for r in var.spec.records : r if r.record_type == "CAA"]
  srv_records   = [for r in var.spec.records : r if r.record_type == "SRV"]
  ptr_records   = [for r in var.spec.records : r if r.record_type == "PTR"]
}

