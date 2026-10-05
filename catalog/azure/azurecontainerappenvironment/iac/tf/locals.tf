locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurecontainerappenvironment"
    "resource_name" = var.metadata.name
  }

  org_tag = var.metadata.org != null ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != null ? { "environment" = var.metadata.env } : {}

  id_tag = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "resource_id" = var.metadata.id } : {}

  # Metadata-derived tags first, then the user's spec tags merged over
  # them: user tags deliberately win so an org's governance conventions
  # can override the derived values where they collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # Logging destination wire value. An explicit choice is honored as-is;
  # unset with a workspace deploys log-analytics (azurerm's own legacy
  # inference); unset without one omits the property (streaming-only).
  logs_destination_map = {
    "LOG_ANALYTICS" = "log-analytics"
    "AZURE_MONITOR" = "azure-monitor"
  }
  logs_destination = (
    var.spec.logs_destination != null
    ? local.logs_destination_map[var.spec.logs_destination]
    : (var.spec.log_analytics_workspace_id != null ? "log-analytics" : null)
  )

  # Public network access wire value; unset lets Azure derive it from
  # the network configuration.
  public_network_access_map = {
    "ENABLED"  = "Enabled"
    "DISABLED" = "Disabled"
  }
  public_network_access = (
    var.spec.public_network_access != null
    ? local.public_network_access_map[var.spec.public_network_access]
    : null
  )

  # Workload profile SKU wire spellings, spelled out row by row so a
  # vocabulary drift fails loudly at plan time instead of deploying a
  # wrong profile.
  workload_profile_type_map = {
    "CONSUMPTION"               = "Consumption"
    "CONSUMPTION_GPU_NC8AS_T4"  = "Consumption-GPU-NC8as-T4"
    "CONSUMPTION_GPU_NC24_A100" = "Consumption-GPU-NC24-A100"
    "D4"                        = "D4"
    "D8"                        = "D8"
    "D16"                       = "D16"
    "D32"                       = "D32"
    "E4"                        = "E4"
    "E8"                        = "E8"
    "E16"                       = "E16"
    "E32"                       = "E32"
    "NC24_A100"                 = "NC24-A100"
    "NC48_A100"                 = "NC48-A100"
    "NC96_A100"                 = "NC96-A100"
  }

  # Managed-identity type wire values.
  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }
}
