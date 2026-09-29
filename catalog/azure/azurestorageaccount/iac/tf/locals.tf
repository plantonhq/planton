locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurestorageaccount"
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
  account_kind_map = {
    "STORAGE_V2"         = "StorageV2"
    "BLOB_STORAGE"       = "BlobStorage"
    "BLOCK_BLOB_STORAGE" = "BlockBlobStorage"
    "FILE_STORAGE"       = "FileStorage"
    "STORAGE"            = "Storage"
  }

  account_tier_map = {
    "STANDARD" = "Standard"
    "PREMIUM"  = "Premium"
  }

  replication_type_map = {
    "LRS"     = "LRS"
    "ZRS"     = "ZRS"
    "GRS"     = "GRS"
    "GZRS"    = "GZRS"
    "RA_GRS"  = "RAGRS"
    "RA_GZRS" = "RAGZRS"
  }

  access_tier_map = {
    "HOT"                 = "Hot"
    "COOL"                = "Cool"
    "COLD"                = "Cold"
    "ACCESS_TIER_PREMIUM" = "Premium"
  }

  # TLS1_2 is the only floor Azure still provisions; the legacy 1.0/1.1
  # floors are retired and no longer exist on the spec enum.
  min_tls_version_map = {
    "TLS1_2" = "TLS1_2"
  }

  allowed_copy_scope_map = {
    "AAD"          = "AAD"
    "PRIVATE_LINK" = "PrivateLink"
  }

  dns_endpoint_type_map = {
    "DNS_ENDPOINT_STANDARD" = "Standard"
    "AZURE_DNS_ZONE"        = "AzureDnsZone"
  }

  encryption_key_type_map = {
    "SERVICE" = "Service"
    "ACCOUNT" = "Account"
  }

  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }

  network_default_action_map = {
    "ALLOW" = "Allow"
    "DENY"  = "Deny"
  }

  network_bypass_map = {
    "AZURE_SERVICES" = "AzureServices"
    "LOGGING"        = "Logging"
    "METRICS"        = "Metrics"
    "NONE"           = "None"
  }

  routing_choice_map = {
    "MICROSOFT_ROUTING" = "MicrosoftRouting"
    "INTERNET_ROUTING"  = "InternetRouting"
  }

  sas_expiration_action_map = {
    "LOG"   = "Log"
    "BLOCK" = "Block"
  }

  immutability_state_map = {
    "DISABLED" = "Disabled"
    "UNLOCKED" = "Unlocked"
    "LOCKED"   = "Locked"
  }

  directory_type_map = {
    "AADDS"   = "AADDS"
    "AADKERB" = "AADKERB"
    "AD"      = "AD"
  }

  default_share_permission_map = {
    "SHARE_PERMISSION_NONE"                 = "None"
    "SHARE_PERMISSION_READER"               = "StorageFileDataSmbShareReader"
    "SHARE_PERMISSION_CONTRIBUTOR"          = "StorageFileDataSmbShareContributor"
    "SHARE_PERMISSION_ELEVATED_CONTRIBUTOR" = "StorageFileDataSmbShareElevatedContributor"
  }

  lifecycle_blob_type_map = {
    "BLOCK_BLOB"  = "blockBlob"
    "APPEND_BLOB" = "appendBlob"
  }

  # Unset enums materialize the spec's documented defaults here (the
  # tfvars wire format drops unspecified enums): kind StorageV2, tier
  # Standard, replication LRS, TLS floor 1.2. access_tier and
  # dns_endpoint_type map unset to null instead -- Azure computes the
  # tier only on the kinds that support it, and the DNS type is a
  # create-only architectural choice azurerm defaults itself.
  account_kind = (
    var.spec.account_kind == null || var.spec.account_kind == "" ? "StorageV2" :
    local.account_kind_map[var.spec.account_kind]
  )

  account_tier = (
    var.spec.account_tier == null || var.spec.account_tier == "" ? "Standard" :
    local.account_tier_map[var.spec.account_tier]
  )

  replication_type = (
    var.spec.replication_type == null || var.spec.replication_type == "" ? "LRS" :
    local.replication_type_map[var.spec.replication_type]
  )

  access_tier = (
    var.spec.access_tier == null || var.spec.access_tier == "" ? null :
    local.access_tier_map[var.spec.access_tier]
  )

  min_tls_version = (
    var.spec.min_tls_version == null || var.spec.min_tls_version == "" ? "TLS1_2" :
    local.min_tls_version_map[var.spec.min_tls_version]
  )

  allowed_copy_scope = (
    var.spec.allowed_copy_scope == null || var.spec.allowed_copy_scope == "" ? null :
    local.allowed_copy_scope_map[var.spec.allowed_copy_scope]
  )

  dns_endpoint_type = (
    var.spec.dns_endpoint_type == null || var.spec.dns_endpoint_type == "" ? null :
    local.dns_endpoint_type_map[var.spec.dns_endpoint_type]
  )

  queue_encryption_key_type = (
    var.spec.queue_encryption_key_type == null || var.spec.queue_encryption_key_type == "" ? null :
    local.encryption_key_type_map[var.spec.queue_encryption_key_type]
  )

  table_encryption_key_type = (
    var.spec.table_encryption_key_type == null || var.spec.table_encryption_key_type == "" ? null :
    local.encryption_key_type_map[var.spec.table_encryption_key_type]
  )

  # provisioned_billing_model_version and edge_zone are sent only when
  # non-empty -- azurerm treats an empty string differently from an
  # omitted argument on ForceNew fields.
  provisioned_billing_model_version = (
    var.spec.provisioned_billing_model_version == null || var.spec.provisioned_billing_model_version == "" ? null :
    var.spec.provisioned_billing_model_version
  )

  edge_zone = (
    var.spec.edge_zone == null || var.spec.edge_zone == "" ? null :
    var.spec.edge_zone
  )

  # Unset network bypass lets azurerm compute Azure's default
  # (AzureServices); an explicit list maps verbatim.
  network_bypass = (
    var.spec.network_rules == null ? null :
    length(var.spec.network_rules.bypass) == 0 ? null :
    [for b in var.spec.network_rules.bypass : local.network_bypass_map[b]]
  )

  # The static website is realized as the standalone
  # azurerm_storage_account_static_website resource (azurerm removed the
  # inline block in v5); empty documents are sent as
  # null so the AtLeastOneOf contract stays visible in the plan.
  static_website_index_document = (
    var.spec.static_website == null ? null :
    var.spec.static_website.index_document == null || var.spec.static_website.index_document == "" ? null :
    var.spec.static_website.index_document
  )

  static_website_error_404_document = (
    var.spec.static_website == null ? null :
    var.spec.static_website.error_404_document == null || var.spec.static_website.error_404_document == "" ? null :
    var.spec.static_website.error_404_document
  )
}
