locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azuremssqlelasticpool"
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

  # Metadata-derived tags first, then the user's spec tags merged over
  # them: user tags deliberately win so an org's governance conventions
  # (cost center, owner) can override the derived values where they
  # collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # The pool addresses its server by name + resource group (azurerm's
  # contract for this resource); both are derived from the server's ARM
  # id so the spec carries ONE authoritative parent reference:
  # /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Sql/servers/{name}
  server_id_parts     = split("/", var.spec.server_id)
  resource_group_name = local.server_id_parts[4]
  server_name         = local.server_id_parts[8]

  # The service tier and hardware family ARM wants alongside the SKU name
  # are pure functions of it -- deriving them makes a name/tier/family
  # mismatch unrepresentable. Both maps are exhaustive over the spec's
  # closed sku_name vocabulary.
  sku_tier_map = {
    "BasicPool"    = "Basic"
    "StandardPool" = "Standard"
    "PremiumPool"  = "Premium"
    "GP_Gen5"      = "GeneralPurpose"
    "GP_Fsv2"      = "GeneralPurpose"
    "GP_DC"        = "GeneralPurpose"
    "BC_Gen5"      = "BusinessCritical"
    "BC_DC"        = "BusinessCritical"
    "HS_Gen5"      = "Hyperscale"
    "HS_PRMS"      = "Hyperscale"
    "HS_MOPRMS"    = "Hyperscale"
  }

  # DTU pools carry no family; vCore pools' family is the name's suffix.
  sku_family_map = {
    "GP_Gen5"   = "Gen5"
    "GP_Fsv2"   = "Fsv2"
    "GP_DC"     = "DC"
    "BC_Gen5"   = "Gen5"
    "BC_DC"     = "DC"
    "HS_Gen5"   = "Gen5"
    "HS_PRMS"   = "PRMS"
    "HS_MOPRMS" = "MOPRMS"
  }

  sku_tier   = local.sku_tier_map[var.spec.sku_name]
  sku_family = lookup(local.sku_family_map, var.spec.sku_name, null)

  enclave_type_map = {
    "VBS"             = "VBS"
    "DEFAULT_ENCLAVE" = "Default"
  }

  license_type_map = {
    "BASE_PRICE"       = "BasePrice"
    "LICENSE_INCLUDED" = "LicenseIncluded"
  }

  enclave_type = (
    var.spec.enclave_type == null || var.spec.enclave_type == "" ? null :
    local.enclave_type_map[var.spec.enclave_type]
  )

  license_type = (
    var.spec.license_type == null || var.spec.license_type == "" ? null :
    local.license_type_map[var.spec.license_type]
  )
}
