locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurepublicipprefix"
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
  # its defaults (IPv4 / Standard / Regional); only an explicit choice is
  # ever sent, so an unspecified spec and Azure's default deploy identically
  # on both engines.
  ip_version = (
    var.spec.ip_version == "IPV4" ? "IPv4" :
    var.spec.ip_version == "IPV6" ? "IPv6" : null
  )
  sku = (
    var.spec.sku == "STANDARD" ? "Standard" :
    var.spec.sku == "STANDARD_V2" ? "StandardV2" : null
  )
  sku_tier = (
    var.spec.sku_tier == "REGIONAL" ? "Regional" :
    var.spec.sku_tier == "GLOBAL" ? "Global" : null
  )
}
