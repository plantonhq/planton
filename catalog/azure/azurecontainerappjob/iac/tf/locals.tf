locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurecontainerappjob"
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

  # Probe transport wire values (Azure validates these case-sensitively).
  probe_transport_map = {
    "TCP_SOCKET" = "TCP"
    "HTTP_GET"   = "HTTP"
    "HTTPS_GET"  = "HTTPS"
  }

  # Volume storage-type wire values; absent deploys EmptyDir.
  volume_storage_type_map = {
    "EMPTY_DIR"      = "EmptyDir"
    "AZURE_FILE"     = "AzureFile"
    "NFS_AZURE_FILE" = "NfsAzureFile"
    "SECRET"         = "Secret"
  }

  # Managed-identity type wire values.
  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }
}
