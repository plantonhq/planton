locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurefrontdoorprofile"
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
    "STANDARD" = "Standard_AzureFrontDoor"
    "PREMIUM"  = "Premium_AzureFrontDoor"
  }
  sku_name = local.sku_map[coalesce(var.spec.sku, "STANDARD")]

  # Identity type: spec enum name -> ARM's value.
  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }

  # Log-scrubbing match variables: spec enum names -> ARM's values.
  log_scrubbing_variable_map = {
    "QUERY_STRING_ARG_NAMES" = "QueryStringArgNames"
    "REQUEST_IP_ADDRESS"     = "RequestIPAddress"
    "REQUEST_URI"            = "RequestUri"
  }
  log_scrubbing_variables = coalesce(var.spec.log_scrubbing_variables, [])
}
