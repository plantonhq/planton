locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurerediscache"
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

  # The spec's sku enum arrives as the FULL proto value name; absent means
  # STANDARD (the tfvars wire format drops zero-valued proto fields, so the
  # module materializes the spec's documented default).
  sku_map = {
    "BASIC"    = "Basic"
    "STANDARD" = "Standard"
    "PREMIUM"  = "Premium"
  }
  sku_name = local.sku_map[coalesce(var.spec.sku_name, "STANDARD")]

  # Azure spells the size as family letter + capacity number; the family
  # letter is fully determined by the tier ("C" for Basic/Standard, "P"
  # for Premium), so the spec never spells it twice.
  family = local.sku_name == "Premium" ? "P" : "C"

  # Patch-schedule days arrive as the spec enum's name string; ARM wants
  # the capitalized English day name.
  day_of_week_map = {
    "MONDAY"    = "Monday"
    "TUESDAY"   = "Tuesday"
    "WEDNESDAY" = "Wednesday"
    "THURSDAY"  = "Thursday"
    "FRIDAY"    = "Friday"
    "SATURDAY"  = "Saturday"
    "SUNDAY"    = "Sunday"
  }

  # Persistence auth method: spec enum name -> ARM's value. null when
  # unset so Azure applies its default (SAS).
  persistence_auth_map = {
    "SAS"              = "SAS"
    "MANAGED_IDENTITY" = "ManagedIdentity"
  }

  # Identity type: spec enum name -> ARM's value.
  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }

  redis_configuration = var.spec.redis_configuration
}
