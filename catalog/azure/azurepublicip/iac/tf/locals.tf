locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurepublicip"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  id_tag = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "resource_id" = var.metadata.id } : {}

  # Metadata-derived tags first, then the user's spec tags merged over them:
  # user tags deliberately win so an org's governance conventions (cost
  # center, owner) can override the derived values where they collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # Map the spec enums' name strings to ARM's values. null lets Azure apply
  # its defaults (Standard / Regional / IPv4 / region-unique label /
  # VirtualNetworkInherited DDoS stance); only an explicit choice is ever
  # sent, so an unspecified spec and Azure's default deploy identically on
  # both engines.
  sku = (
    var.spec.sku == "STANDARD" ? "Standard" :
    var.spec.sku == "STANDARD_V2" ? "StandardV2" : null
  )
  sku_tier = (
    var.spec.sku_tier == "REGIONAL" ? "Regional" :
    var.spec.sku_tier == "GLOBAL" ? "Global" : null
  )
  ip_version = (
    var.spec.ip_version == "IPV4" ? "IPv4" :
    var.spec.ip_version == "IPV6" ? "IPv6" : null
  )
  domain_name_label_scope = (
    var.spec.domain_name_label_scope == "TENANT_REUSE" ? "TenantReuse" :
    var.spec.domain_name_label_scope == "SUBSCRIPTION_REUSE" ? "SubscriptionReuse" :
    var.spec.domain_name_label_scope == "RESOURCE_GROUP_REUSE" ? "ResourceGroupReuse" :
    var.spec.domain_name_label_scope == "NO_REUSE" ? "NoReuse" : null
  )
  ddos_protection_mode = (
    var.spec.ddos_protection_mode == "DISABLED" ? "Disabled" :
    var.spec.ddos_protection_mode == "ENABLED" ? "Enabled" : null
  )
}
