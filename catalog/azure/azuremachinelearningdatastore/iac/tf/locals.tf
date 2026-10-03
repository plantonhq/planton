locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azuremachinelearningdatastore"
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
  # NOTE: datastore tags are ForceNew on the provider -- changing any tag
  # replaces the datastore object (the data it points at is untouched).
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # The spec's service-data identity modes to the provider's wire values.
  # Unspecified (the enum's zero value renders as "") maps to null so the
  # provider applies its default, "None". The provider names this argument
  # `service_data_auth_identity` on the blob resource and
  # `service_data_identity` on the other two -- ONE spec field feeds both
  # (recorded in the parity manifest).
  service_data_identity_wire = {
    "NONE"                               = "None"
    "WORKSPACE_SYSTEM_ASSIGNED_IDENTITY" = "WorkspaceSystemAssignedIdentity"
    "WORKSPACE_USER_ASSIGNED_IDENTITY"   = "WorkspaceUserAssignedIdentity"
  }
}
