locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azureeventhubnamespace"
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
  # value name; an absent sku deploys STANDARD -- the full-featured
  # multi-tenant tier that fits most workloads.
  sku_map = {
    "BASIC"    = "Basic"
    "STANDARD" = "Standard"
    "PREMIUM"  = "Premium"
  }
  sku = (
    var.spec.sku != null && var.spec.sku != ""
    ? local.sku_map[var.spec.sku]
    : "Standard"
  )

  # Identity type wire values -- Event Hubs namespaces support all three
  # managed-identity models.
  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }

  # Firewall default-action wire values -- Azure requires an explicit
  # choice when the rule set is declared (the spec enum rejects
  # unspecified).
  network_default_action_map = {
    "ALLOW" = "Allow"
    "DENY"  = "Deny"
  }
}
