locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azuremssqldatabase"
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
    "DEFAULT"                            = "Default"
    "COPY"                               = "Copy"
    "SECONDARY"                          = "Secondary"
    "ONLINE_SECONDARY"                   = "OnlineSecondary"
    "POINT_IN_TIME_RESTORE"              = "PointInTimeRestore"
    "RECOVERY"                           = "Recovery"
    "RESTORE"                            = "Restore"
    "RESTORE_LONG_TERM_RETENTION_BACKUP" = "RestoreLongTermRetentionBackup"
  }

  secondary_type_map = {
    "GEO"   = "Geo"
    "NAMED" = "Named"
  }

  license_type_map = {
    "BASE_PRICE"       = "BasePrice"
    "LICENSE_INCLUDED" = "LicenseIncluded"
  }

  enclave_type_map = {
    "VBS"             = "VBS"
    "DEFAULT_ENCLAVE" = "Default"
  }

  storage_account_type_map = {
    "GEO_REDUNDANT"      = "Geo"
    "GEO_ZONE_REDUNDANT" = "GeoZone"
    "LOCALLY_REDUNDANT"  = "Local"
    "ZONE_REDUNDANT"     = "Zone"
  }

  import_storage_key_type_map = {
    "SHARED_ACCESS_KEY"  = "SharedAccessKey"
    "STORAGE_ACCESS_KEY" = "StorageAccessKey"
  }

  import_authentication_type_map = {
    "SQL"         = "Sql"
    "AD_PASSWORD" = "ADPassword"
  }

  threat_detection_state_map = {
    "ENABLED"  = "Enabled"
    "DISABLED" = "Disabled"
  }

  # Unset enums map to null so azurerm applies (or computes) its own
  # default instead of the module materializing a value.
  create_mode = (
    var.spec.create_mode == null || var.spec.create_mode == "" ? null :
    local.create_mode_map[var.spec.create_mode]
  )

  secondary_type = (
    var.spec.secondary_type == null || var.spec.secondary_type == "" ? null :
    local.secondary_type_map[var.spec.secondary_type]
  )

  license_type = (
    var.spec.license_type == null || var.spec.license_type == "" ? null :
    local.license_type_map[var.spec.license_type]
  )

  enclave_type = (
    var.spec.enclave_type == null || var.spec.enclave_type == "" ? null :
    local.enclave_type_map[var.spec.enclave_type]
  )

  storage_account_type = (
    var.spec.storage_account_type == null || var.spec.storage_account_type == "" ? null :
    local.storage_account_type_map[var.spec.storage_account_type]
  )

  sku_name = var.spec.sku_name == null || var.spec.sku_name == "" ? null : var.spec.sku_name
}
