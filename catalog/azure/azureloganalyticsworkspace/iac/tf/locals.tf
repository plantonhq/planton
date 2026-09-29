locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azureloganalyticsworkspace"
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

  id_tag = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "resource_id" = var.metadata.id } : {}

  # Merge base, org, environment, id, and user tags -- user tags win on key
  # conflicts (the governance surface belongs to the user).
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # SKU wire values. The tfvars wire format carries the FULL proto enum
  # value name; an absent sku deploys Azure's recommended PerGB2018.
  # Standard/Premium/LACluster/Unlimited are deliberately not mapped:
  # Azure blocks creating workspaces on them (see the spec enum comment).
  sku_map = {
    "PER_GB_2018"          = "PerGB2018"
    "CAPACITY_RESERVATION" = "CapacityReservation"
    "PER_NODE"             = "PerNode"
    "STANDALONE"           = "Standalone"
  }
  sku = (
    var.spec.sku != null && var.spec.sku != ""
    ? local.sku_map[var.spec.sku]
    : "PerGB2018"
  )

  # Identity type wire values. Workspaces accept exactly SystemAssigned or
  # UserAssigned -- the combined model does not exist on this resource.
  identity_type_map = {
    "SYSTEM_ASSIGNED" = "SystemAssigned"
    "USER_ASSIGNED"   = "UserAssigned"
  }
}
