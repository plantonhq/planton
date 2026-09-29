locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azuremysqlflexibleserver"
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

  # Metadata-derived tags first, then the user's spec tags merged over
  # them: user tags deliberately win so an org's governance conventions
  # (cost center, owner) can override the derived values where they
  # collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # The spec's enums arrive as FULL proto value names (the tfvars wire
  # format never strips prefixes); each map below carries the complete
  # verbatim vocabulary for its enum, mapped to ARM's values. A missing
  # entry would silently drop the setting, so the maps are exhaustive by
  # construction.
  create_mode_map = {
    "DEFAULT"               = "Default"
    "POINT_IN_TIME_RESTORE" = "PointInTimeRestore"
    "REPLICA"               = "Replica"
    "GEO_RESTORE"           = "GeoRestore"
  }

  ha_mode_map = {
    "ZONE_REDUNDANT" = "ZoneRedundant"
    "SAME_ZONE"      = "SameZone"
  }

  public_network_access_map = {
    "ENABLED"  = "Enabled"
    "DISABLED" = "Disabled"
  }

  # Unspecified create_mode means a fresh (DEFAULT) server. azurerm treats
  # an omitted create_mode the same way, so unset maps to null rather than
  # materializing "Default" -- keeping the plan free of a no-op diff.
  create_mode = (
    var.spec.create_mode == null || var.spec.create_mode == "" ? null :
    local.create_mode_map[var.spec.create_mode]
  )

  is_default_mode = local.create_mode == null || local.create_mode == "Default"

  # version and sku_name are inherited from the source on replicas and
  # restores; sending them would fight the service, so they are only
  # forwarded for a fresh server (version) or when explicitly set (sku --
  # a replica may legitimately override its compute).
  version = local.is_default_mode ? coalesce(var.spec.version, "8.0.21") : null

  sku_name = var.spec.sku_name == null || var.spec.sku_name == "" ? null : var.spec.sku_name

  # Credentials are sent only when non-empty: replicas and restores
  # inherit them from the source, and azurerm rejects empty strings.
  administrator_login = (
    var.spec.administrator_login == null || var.spec.administrator_login == "" ? null :
    var.spec.administrator_login
  )

  administrator_password = (
    var.spec.administrator_password == null || var.spec.administrator_password == "" ? null :
    var.spec.administrator_password
  )

  # replication_role can never be part of a create call (Azure rejects
  # it); azurerm applies it as a day-2 update, which is exactly how a
  # replica is promoted: flip the field on the existing replica and apply.
  replication_role = (
    var.spec.replication_role == null || var.spec.replication_role == "" ? null : "None"
  )

  # Unset public_network_access maps to null so Azure derives the value
  # (Enabled for a public server, Disabled when VNet-injected) instead of
  # the module guessing and fighting the service.
  public_network_access = (
    var.spec.public_network_access == null || var.spec.public_network_access == "" ? null :
    local.public_network_access_map[var.spec.public_network_access]
  )

  # The Entra tenant for the AAD administrator grant: the spec's explicit
  # tenant wins, otherwise the deploying credential's tenant -- the
  # correct value for virtually every deployment.
  aad_tenant_id = (
    var.spec.aad_administrator != null && var.spec.aad_administrator.tenant_id != null && var.spec.aad_administrator.tenant_id != ""
    ? var.spec.aad_administrator.tenant_id
    : data.azurerm_client_config.current.tenant_id
  )
}
