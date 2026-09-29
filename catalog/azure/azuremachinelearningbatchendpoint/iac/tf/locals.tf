locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azuremachinelearningbatchendpoint"
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

  # The spec's identity flavors to the provider's wire values. azapi
  # accepts the azurerm-style comma+space dual value and normalizes it to
  # ARM's own "SystemAssigned,UserAssigned" on the wire -- the same ARM
  # identity type the Pulumi module sends.
  identity_type_wire = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }

  # The ARM properties object, assembled key-by-key so unset optionals are
  # OMITTED (ARM applies its own defaults) instead of sent as nulls.
  # authMode always sends: ARM requires it, and AADToken is the only mode
  # the batch service accepts (the spec's vocabulary already enforces it;
  # the unset default applies here). Unlike the online endpoint there is
  # NO publicNetworkAccess property on the batch surface -- reachability
  # is governed by the workspace's own network settings.
  endpoint_properties = merge(
    {
      authMode = var.spec.auth_mode != "" ? var.spec.auth_mode : "AADToken"
    },
    var.spec.default_deployment_name != "" ? {
      defaults = { deploymentName = var.spec.default_deployment_name }
    } : {},
    var.spec.description != "" ? { description = var.spec.description } : {},
    length(var.spec.properties) > 0 ? { properties = var.spec.properties } : {},
  )
}
