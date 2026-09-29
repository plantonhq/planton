locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azureexpressroutecircuit"
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

  # The spec's enum NAMES mapped onto ARM's SKU vocabulary. Tier and
  # family combine into the ARM SKU name ("Standard_MeteredData") on the
  # provider side.
  sku_tier_wire = {
    "BASIC"    = "Basic"
    "LOCAL"    = "Local"
    "STANDARD" = "Standard"
    "PREMIUM"  = "Premium"
  }
  sku_family_wire = {
    "METERED_DATA"   = "MeteredData"
    "UNLIMITED_DATA" = "UnlimitedData"
  }
}
