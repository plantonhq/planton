locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurepostgresqlflexibleserver"
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
    "REVIVE_DROPPED"        = "ReviveDropped"
  }

  ha_mode_map = {
    "ZONE_REDUNDANT" = "ZoneRedundant"
    "SAME_ZONE"      = "SameZone"
  }

  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }

  principal_type_map = {
    "USER"              = "User"
    "GROUP"             = "Group"
    "SERVICE_PRINCIPAL" = "ServicePrincipal"
  }

  # Storage tier names are identical on both sides -- the map exists so an
  # unknown value fails the plan loudly instead of passing through as an
  # unmapped string.
  storage_tier_map = {
    "P4"  = "P4"
    "P6"  = "P6"
    "P10" = "P10"
    "P15" = "P15"
    "P20" = "P20"
    "P30" = "P30"
    "P40" = "P40"
    "P50" = "P50"
    "P60" = "P60"
    "P70" = "P70"
    "P80" = "P80"
  }

  # Unspecified create_mode means a fresh (DEFAULT) server. azurerm treats
  # an omitted create_mode the same way, so unset maps to null rather than
  # materializing "Default" -- keeping the plan free of a no-op diff.
  create_mode = (
    var.spec.create_mode == null || var.spec.create_mode == "" ? null :
    local.create_mode_map[var.spec.create_mode]
  )

  is_default_mode = local.create_mode == null || local.create_mode == "Default"

  # version and sku_name/storage are inherited from the source on
  # replicas and restores; sending them would fight the service, so they
  # are only forwarded for a fresh server (version) or when explicitly set
  # (sku/storage -- a replica may legitimately override its compute).
  version = local.is_default_mode ? coalesce(var.spec.version, "16") : null

  sku_name = var.spec.sku_name == null || var.spec.sku_name == "" ? null : var.spec.sku_name

  storage_tier = (
    var.spec.storage_tier == null || var.spec.storage_tier == "" ? null :
    local.storage_tier_map[var.spec.storage_tier]
  )

  # Password-auth credentials are sent only when non-empty: an Entra-only
  # server (password auth disabled) and a replica both legitimately omit
  # them, and azurerm rejects empty strings.
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

  # Whether Entra (AAD) auth is enabled -- drives the tenant fallback and
  # the AAD administrator sub-resources.
  aad_auth_enabled = (
    var.spec.authentication != null && var.spec.authentication.active_directory_auth_enabled == true
  )

  # The Entra tenant for AAD auth and administrator grants: the spec's
  # explicit tenant wins, otherwise the deploying credential's tenant --
  # the correct value for virtually every deployment.
  aad_tenant_id = (
    var.spec.authentication != null && var.spec.authentication.tenant_id != null && var.spec.authentication.tenant_id != ""
    ? var.spec.authentication.tenant_id
    : data.azurerm_client_config.current.tenant_id
  )

  # Whether the identity block includes a system-assigned identity, which
  # is when the principal-id output is populated.
  has_system_identity = (
    var.spec.identity != null && contains(["SYSTEM_ASSIGNED", "SYSTEM_AND_USER_ASSIGNED"], var.spec.identity.type)
  )
}
