locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurecosmosdbaccount"
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
  # verbatim vocabulary for its enum, mapped to ARM's wire values. A
  # missing entry would silently drop the setting, so the maps are
  # exhaustive by construction.
  kind_map = {
    "GLOBAL_DOCUMENT_DB" = "GlobalDocumentDB"
    "MONGO_DB"           = "MongoDB"
  }

  # Unspecified kind means the SQL (NoSQL) API -- azurerm's own default.
  kind = (
    var.spec.kind == null || var.spec.kind == "" ? "GlobalDocumentDB" :
    local.kind_map[var.spec.kind]
  )

  consistency_level_map = {
    "STRONG"            = "Strong"
    "BOUNDED_STALENESS" = "BoundedStaleness"
    "SESSION"           = "Session"
    "CONSISTENT_PREFIX" = "ConsistentPrefix"
    "EVENTUAL"          = "Eventual"
  }

  # Unspecified consistency means Session -- Azure's recommended default.
  consistency_level = (
    var.spec.consistency_policy.consistency_level == null || var.spec.consistency_policy.consistency_level == ""
    ? "Session"
    : local.consistency_level_map[var.spec.consistency_policy.consistency_level]
  )

  # The staleness dials only mean something on BoundedStaleness; azurerm
  # diff-suppresses them elsewhere, so they are passed through with the
  # proto defaults (5 / 100) materialized when unset.
  max_interval_in_seconds = (
    var.spec.consistency_policy.max_interval_in_seconds != null
    ? var.spec.consistency_policy.max_interval_in_seconds
    : 5
  )
  max_staleness_prefix = (
    var.spec.consistency_policy.max_staleness_prefix != null
    ? var.spec.consistency_policy.max_staleness_prefix
    : 100
  )

  capability_map = {
    "ENABLE_SERVERLESS"                      = "EnableServerless"
    "ENABLE_CASSANDRA"                       = "EnableCassandra"
    "ENABLE_GREMLIN"                         = "EnableGremlin"
    "ENABLE_TABLE"                           = "EnableTable"
    "ENABLE_AGGREGATION_PIPELINE"            = "EnableAggregationPipeline"
    "ENABLE_MONGO"                           = "EnableMongo"
    "ENABLE_MONGO_16MB_DOCUMENT_SUPPORT"     = "EnableMongo16MBDocumentSupport"
    "MONGO_DB_V34"                           = "MongoDBv3.4"
    "MONGO_ENABLE_DOC_LEVEL_TTL"             = "mongoEnableDocLevelTTL"
    "DELETE_ALL_ITEMS_BY_PARTITION_KEY"      = "DeleteAllItemsByPartitionKey"
    "DISABLE_RATE_LIMITING_RESPONSES"        = "DisableRateLimitingResponses"
    "ALLOW_SELF_SERVE_UPGRADE_TO_MONGO36"    = "AllowSelfServeUpgradeToMongo36"
    "ENABLE_MONGO_RETRYABLE_WRITES"          = "EnableMongoRetryableWrites"
    "ENABLE_MONGO_ROLE_BASED_ACCESS_CONTROL" = "EnableMongoRoleBasedAccessControl"
    "ENABLE_UNIQUE_COMPOUND_NESTED_DOCS"     = "EnableUniqueCompoundNestedDocs"
    "ENABLE_NO_SQL_VECTOR_SEARCH"            = "EnableNoSQLVectorSearch"
    "ENABLE_NO_SQL_FULL_TEXT_SEARCH"         = "EnableNoSQLFullTextSearch"
    "ENABLE_TTL_ON_CUSTOM_PATH"              = "EnableTtlOnCustomPath"
    "ENABLE_PARTIAL_UNIQUE_INDEX"            = "EnablePartialUniqueIndex"
    "ENABLE_FABRIC_NETWORK_ACL_BYPASS"       = "EnableFabricNetworkAclBypass"
  }

  # Capabilities are exactly what the spec declares -- the module never
  # injects one silently (a MONGO_DB account declares ENABLE_MONGO
  # itself; presets and docs teach this).
  capabilities = [for c in var.spec.capabilities : local.capability_map[c]]

  backup_type_map = {
    "PERIODIC"   = "Periodic"
    "CONTINUOUS" = "Continuous"
  }

  continuous_tier_map = {
    "CONTINUOUS_7_DAYS"  = "Continuous7Days"
    "CONTINUOUS_30_DAYS" = "Continuous30Days"
  }

  backup_storage_redundancy_map = {
    "GEO"   = "Geo"
    "LOCAL" = "Local"
    "ZONE"  = "Zone"
  }

  mongo_server_version_map = {
    "MONGO_3_2" = "3.2"
    "MONGO_3_6" = "3.6"
    "MONGO_4_0" = "4.0"
    "MONGO_4_2" = "4.2"
    "MONGO_5_0" = "5.0"
    "MONGO_6_0" = "6.0"
    "MONGO_7_0" = "7.0"
  }

  mongo_server_version = (
    var.spec.mongo_server_version == null || var.spec.mongo_server_version == ""
    ? null
    : local.mongo_server_version_map[var.spec.mongo_server_version]
  )

  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }

  # The default identity is a composite wire string: the plain name for
  # first-party/system, or "UserAssignedIdentity=<ARM id>" carrying the
  # identity the account acts as. Composed here so the spec can model
  # the identity as a real reference instead of a magic string.
  default_identity_type = (
    var.spec.default_identity == null ? null :
    var.spec.default_identity.type == "FIRST_PARTY" ? "FirstPartyIdentity" :
    var.spec.default_identity.type == "SYSTEM_ASSIGNED_DEFAULT" ? "SystemAssignedIdentity" :
    "UserAssignedIdentity=${var.spec.default_identity.user_assigned_identity_id}"
  )

  analytical_schema_type_map = {
    "WELL_DEFINED"  = "WellDefined"
    "FULL_FIDELITY" = "FullFidelity"
  }

  # Tls12 is the only floor Azure still provisions; the legacy Tls/Tls11
  # floors are retired and no longer exist on the spec enum.
  minimal_tls_version_map = {
    "TLS_1_2" = "Tls12"
  }

  # Unset means Azure's own TLS 1.2 default -- materialized explicitly
  # so both engines send the same value.
  minimal_tls_version = (
    var.spec.minimal_tls_version == null || var.spec.minimal_tls_version == ""
    ? "Tls12"
    : local.minimal_tls_version_map[var.spec.minimal_tls_version]
  )

  create_mode_map = {
    "DEFAULT" = "Default"
    "RESTORE" = "Restore"
  }

  # Sent only when the spec sets it: azurerm rejects create_mode on
  # accounts without continuous backup, and the spec-level rule mirrors
  # that contract.
  create_mode = (
    var.spec.create_mode == null || var.spec.create_mode == ""
    ? null
    : local.create_mode_map[var.spec.create_mode]
  )
}
