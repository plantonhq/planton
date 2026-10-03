locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azuremonitoractivitylogalert"
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

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # --- Enum name -> ARM value maps (mirror the Pulumi module's mappers) ---

  location_map = {
    "GLOBAL"         = "global"
    "WEST_EUROPE"    = "westeurope"
    "NORTH_EUROPE"   = "northeurope"
    "EAST_US_2_EUAP" = "eastus2euap"
  }
  location = lookup(local.location_map, coalesce(var.spec.location, "GLOBAL"), "global")

  category_map = {
    "ADMINISTRATIVE"  = "Administrative"
    "AUTOSCALE"       = "Autoscale"
    "POLICY"          = "Policy"
    "RECOMMENDATION"  = "Recommendation"
    "RESOURCE_HEALTH" = "ResourceHealth"
    "SECURITY"        = "Security"
    "SERVICE_HEALTH"  = "ServiceHealth"
  }
  category = lookup(local.category_map, var.spec.criteria.category, null)

  level_map = {
    "VERBOSE"       = "Verbose"
    "INFORMATIONAL" = "Informational"
    "WARNING"       = "Warning"
    "ERROR"         = "Error"
    "CRITICAL"      = "Critical"
  }
  levels = [for l in var.spec.criteria.levels : local.level_map[l]]

  recommendation_category_map = {
    "COST"                    = "Cost"
    "RELIABILITY"             = "Reliability"
    "OPERATIONAL_EXCELLENCE"  = "OperationalExcellence"
    "PERFORMANCE"             = "Performance"
    "HIGH_AVAILABILITY"       = "HighAvailability"
    "SECURITY_RECOMMENDATION" = "Security"
  }
  recommendation_category = var.spec.criteria.recommendation_category != "" ? local.recommendation_category_map[var.spec.criteria.recommendation_category] : null

  recommendation_impact_map = {
    "HIGH"   = "High"
    "MEDIUM" = "Medium"
    "LOW"    = "Low"
  }
  recommendation_impact = var.spec.criteria.recommendation_impact != "" ? local.recommendation_impact_map[var.spec.criteria.recommendation_impact] : null

  health_status_map = {
    "AVAILABLE"   = "Available"
    "DEGRADED"    = "Degraded"
    "UNAVAILABLE" = "Unavailable"
    "UNKNOWN"     = "Unknown"
  }
  health_reason_map = {
    "PLATFORM_INITIATED" = "PlatformInitiated"
    "USER_INITIATED"     = "UserInitiated"
    "REASON_UNKNOWN"     = "Unknown"
  }
  service_health_event_map = {
    "INCIDENT"            = "Incident"
    "MAINTENANCE"         = "Maintenance"
    "EVENT_INFORMATIONAL" = "Informational"
    "ACTION_REQUIRED"     = "ActionRequired"
    "EVENT_SECURITY"      = "Security"
  }

  resource_health = var.spec.criteria.resource_health == null ? null : {
    current  = [for s in var.spec.criteria.resource_health.current : local.health_status_map[s]]
    previous = [for s in var.spec.criteria.resource_health.previous : local.health_status_map[s]]
    reason   = [for r in var.spec.criteria.resource_health.reason : local.health_reason_map[r]]
  }

  service_health = var.spec.criteria.service_health == null ? null : {
    events    = [for e in var.spec.criteria.service_health.events : local.service_health_event_map[e]]
    locations = var.spec.criteria.service_health.locations
    services  = var.spec.criteria.service_health.services
  }
}
