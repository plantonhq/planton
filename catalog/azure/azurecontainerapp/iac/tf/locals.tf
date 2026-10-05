locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurecontainerapp"
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

  # Revision mode wire value; absent deploys Single.
  revision_mode_map = {
    "SINGLE"   = "Single"
    "MULTIPLE" = "Multiple"
  }
  revision_mode = var.spec.revision_mode != null ? local.revision_mode_map[var.spec.revision_mode] : "Single"

  # Probe transport wire values (Azure validates these case-sensitively).
  probe_transport_map = {
    "TCP_SOCKET" = "TCP"
    "HTTP_GET"   = "HTTP"
    "HTTPS_GET"  = "HTTPS"
  }

  # Ingress transport wire values; absent deploys auto.
  ingress_transport_map = {
    "AUTO"  = "auto"
    "HTTP"  = "http"
    "HTTP2" = "http2"
    "TCP"   = "tcp"
  }

  # mTLS client-certificate mode wire values; absent is never sent.
  # ARM's wire vocabulary here is lowercase (accept/require/ignore) --
  # unlike most ARM enums; the SDK's Go identifiers are capitalized but
  # the string constants are not.
  client_certificate_mode_map = {
    "ACCEPT"  = "accept"
    "REQUIRE" = "require"
    "IGNORE"  = "ignore"
  }

  # IP restriction action wire values.
  ip_restriction_action_map = {
    "ALLOW" = "Allow"
    "DENY"  = "Deny"
  }

  # Volume storage-type wire values; absent deploys EmptyDir.
  volume_storage_type_map = {
    "EMPTY_DIR"      = "EmptyDir"
    "AZURE_FILE"     = "AzureFile"
    "NFS_AZURE_FILE" = "NfsAzureFile"
    "SECRET"         = "Secret"
  }

  # Dapr app-protocol wire values; absent deploys http.
  dapr_protocol_map = {
    "DAPR_HTTP" = "http"
    "DAPR_GRPC" = "grpc"
  }

  # Managed-identity type wire values.
  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }
}
