locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurelinuxwebapp"
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

  # Merge Application Insights connection string into app settings --
  # the connection-string form is how the App Service agent discovers
  # the telemetry destination.
  ai_settings = var.spec.application_insights_connection_string != null ? {
    APPLICATIONINSIGHTS_CONNECTION_STRING = var.spec.application_insights_connection_string
  } : {}

  merged_app_settings = merge(
    var.spec.app_settings,
    local.ai_settings,
  )

  # Spec enum name strings -> Azure's wire values. Spelled out row by
  # row so the plan renders exact wire values and a vocabulary drift
  # fails loudly at plan time.

  client_certificate_mode_map = {
    "REQUIRED"                  = "Required"
    "OPTIONAL"                  = "Optional"
    "OPTIONAL_INTERACTIVE_USER" = "OptionalInteractiveUser"
  }
  client_certificate_mode = local.client_certificate_mode_map[coalesce(var.spec.client_certificate_mode, "OPTIONAL")]

  tls_version_map = {
    "TLS_1_0" = "1.0"
    "TLS_1_1" = "1.1"
    "TLS_1_2" = "1.2"
    "TLS_1_3" = "1.3"
  }
  minimum_tls_version     = local.tls_version_map[coalesce(var.spec.site_config.minimum_tls_version, "TLS_1_2")]
  scm_minimum_tls_version = local.tls_version_map[coalesce(var.spec.site_config.scm_minimum_tls_version, "TLS_1_2")]

  ftps_state_map = {
    "ALL_ALLOWED" = "AllAllowed"
    "FTPS_ONLY"   = "FtpsOnly"
    "DISABLED"    = "Disabled"
  }
  ftps_state = local.ftps_state_map[coalesce(var.spec.site_config.ftps_state, "DISABLED")]

  load_balancing_mode_map = {
    "LEAST_REQUESTS"         = "LeastRequests"
    "WEIGHTED_ROUND_ROBIN"   = "WeightedRoundRobin"
    "LEAST_RESPONSE_TIME"    = "LeastResponseTime"
    "WEIGHTED_TOTAL_TRAFFIC" = "WeightedTotalTraffic"
    "REQUEST_HASH"           = "RequestHash"
    "PER_SITE_ROUND_ROBIN"   = "PerSiteRoundRobin"
  }
  load_balancing_mode = local.load_balancing_mode_map[coalesce(var.spec.site_config.load_balancing_mode, "LEAST_REQUESTS")]

  managed_pipeline_mode_map = {
    "INTEGRATED" = "Integrated"
    "CLASSIC"    = "Classic"
  }
  managed_pipeline_mode = local.managed_pipeline_mode_map[coalesce(var.spec.site_config.managed_pipeline_mode, "INTEGRATED")]

  ip_restriction_action_map = {
    "ALLOW" = "Allow"
    "DENY"  = "Deny"
  }
  ip_restriction_default_action     = local.ip_restriction_action_map[coalesce(var.spec.site_config.ip_restriction_default_action, "ALLOW")]
  scm_ip_restriction_default_action = local.ip_restriction_action_map[coalesce(var.spec.site_config.scm_ip_restriction_default_action, "ALLOW")]

  java_server_map = {
    "JAVA_SE"  = "JAVA"
    "TOMCAT"   = "TOMCAT"
    "JBOSSEAP" = "JBOSSEAP"
  }

  connection_string_type_map = {
    "MYSQL"            = "MySQL"
    "SQL_SERVER"       = "SQLServer"
    "SQL_AZURE"        = "SQLAzure"
    "CUSTOM"           = "Custom"
    "NOTIFICATION_HUB" = "NotificationHub"
    "SERVICE_BUS"      = "ServiceBus"
    "EVENT_HUB"        = "EventHub"
    "API_HUB"          = "APIHub"
    "DOC_DB"           = "DocDb"
    "REDIS_CACHE"      = "RedisCache"
    "POSTGRESQL"       = "PostgreSQL"
  }

  storage_mount_type_map = {
    "AZURE_FILES" = "AzureFiles"
    "AZURE_BLOB"  = "AzureBlob"
  }

  log_level_map = {
    "OFF"         = "Off"
    "ERROR"       = "Error"
    "WARNING"     = "Warning"
    "INFORMATION" = "Information"
    "VERBOSE"     = "Verbose"
  }

  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }

  backup_frequency_unit_map = {
    "DAY"  = "Day"
    "HOUR" = "Hour"
  }

  unauthenticated_action_map = {
    "REDIRECT_TO_LOGIN_PAGE" = "RedirectToLoginPage"
    "ALLOW_ANONYMOUS"        = "AllowAnonymous"
    "RETURN_401"             = "Return401"
    "RETURN_403"             = "Return403"
  }

  forward_proxy_convention_map = {
    "FORWARD_PROXY_NO_PROXY" = "NoProxy"
    "FORWARD_PROXY_STANDARD" = "Standard"
    "FORWARD_PROXY_CUSTOM"   = "Custom"
  }

  cookie_expiration_convention_map = {
    "FIXED_TIME"                = "FixedTime"
    "IDENTITY_PROVIDER_DERIVED" = "IdentityProviderDerived"
  }
}
