locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurecontainerregistry"
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

  # Map the spec enum's name string to ARM's SKU value. The unspecified
  # spec applies the STANDARD baseline on both engines (azurerm requires
  # an explicit sku, so the default is materialized here rather than sent
  # as null).
  sku = (
    var.spec.sku == "BASIC" ? "Basic" :
    var.spec.sku == "PREMIUM" ? "Premium" : "Standard"
  )

  # Map the bypass enum's name string to ARM's value. null lets Azure
  # apply its default (AzureServices); only an explicit choice is sent, so
  # an unspecified spec and Azure's default deploy identically on both
  # engines.
  network_rule_bypass_option = (
    var.spec.network_rule_bypass_option == "AZURE_SERVICES" ? "AzureServices" :
    var.spec.network_rule_bypass_option == "NONE" ? "None" : null
  )

  # Map the identity type enum's name string to ARM's comma-separated
  # value.
  identity_type = (
    var.spec.identity == null ? null :
    var.spec.identity.type == "SYSTEM_ASSIGNED" ? "SystemAssigned" :
    var.spec.identity.type == "USER_ASSIGNED" ? "UserAssigned" :
    var.spec.identity.type == "SYSTEM_AND_USER_ASSIGNED" ? "SystemAssigned, UserAssigned" : null
  )

  # Map the network-rule default action enum's name string to ARM's value.
  # Azure's default (Allow) applies when unspecified.
  network_rule_default_action = (
    var.spec.network_rule_set == null ? null :
    var.spec.network_rule_set.default_action == "DENY" ? "Deny" : "Allow"
  )
}
