locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciDnsZone"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  zone_type_map = {
    "primary"   = "PRIMARY"
    "secondary" = "SECONDARY"
  }

  scope_value = (
    var.spec.scope == "private" ? "PRIVATE" :
    var.spec.scope == "global" ? "GLOBAL" :
    null
  )

  dnssec_state = (
    var.spec.is_dnssec_enabled == true ? "ENABLED" :
    var.spec.is_dnssec_enabled == false ? "DISABLED" :
    null
  )
}
