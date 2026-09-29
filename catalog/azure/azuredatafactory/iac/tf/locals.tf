locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azuredatafactory"
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

  # The spec enum's value names mapped to the provider's identity tokens.
  # Like the Event Grid namespace, a factory supports the combined mode.
  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }

  # Composed children keyed by name -- renames replace only that child,
  # siblings stay untouched, in lockstep with the Pulumi module's
  # per-name resources. Credential names share one namespace under the
  # factory (spec CEL enforces cross-list uniqueness).
  user_managed_identity_credentials_by_name = {
    for credential in var.spec.user_managed_identity_credentials : credential.name => credential
  }

  service_principal_credentials_by_name = {
    for credential in var.spec.service_principal_credentials : credential.name => credential
  }

  managed_private_endpoints_by_name = {
    for endpoint in var.spec.managed_private_endpoints : endpoint.name => endpoint
  }
}
