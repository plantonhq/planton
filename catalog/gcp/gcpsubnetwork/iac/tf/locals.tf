locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain; an empty string would be sent
  # verbatim and rejected by the API.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Planton middleware applies the spec's proto defaults before the module
  # runs; the coalesces here are only a safety net for direct tfvars
  # invocations so behavior stays identical.
  purpose    = coalesce(var.spec.purpose, "PRIVATE")
  stack_type = coalesce(var.spec.stack_type, "IPV4_ONLY")

  # Normalize "" -> null for optionals the provider treats as meaningfully
  # absent (an empty role or ipv6_access_type would be rejected).
  ip_cidr_range              = var.spec.ip_cidr_range != "" ? var.spec.ip_cidr_range : null
  role                       = var.spec.role != "" ? var.spec.role : null
  ipv6_access_type           = var.spec.ipv6_access_type != "" ? var.spec.ipv6_access_type : null
  external_ipv6_prefix       = var.spec.external_ipv6_prefix != "" ? var.spec.external_ipv6_prefix : null
  private_ipv6_google_access = var.spec.private_ipv6_google_access != "" ? var.spec.private_ipv6_google_access : null

  private_ip_google_access = coalesce(var.spec.private_ip_google_access, false)

  # Drop entries the tfvars converter may emit as empty objects so the API
  # never sees a blank secondary range. A range is real when it has a name
  # and either CIDR source (literal or reserved internal range).
  secondary_ip_ranges = var.spec.secondary_ip_ranges != null ? [
    for secondary_range in var.spec.secondary_ip_ranges : secondary_range
    if secondary_range.range_name != "" && (secondary_range.ip_cidr_range != "" || secondary_range.reserved_internal_range != "")
  ] : []

  # Flow-log defaults mirror the GCP API's own (5s aggregation, 50% sampling,
  # all metadata, no filter) so an empty log_config object turns logging on
  # with sane behavior — identical to the Pulumi module.
  log_config = var.spec.log_config == null || !coalesce(var.spec.log_config.enabled, true) ? null : {
    aggregation_interval = coalesce(var.spec.log_config.aggregation_interval, "INTERVAL_5_SEC")
    flow_sampling        = coalesce(var.spec.log_config.flow_sampling, 0.5)
    metadata             = coalesce(var.spec.log_config.metadata, "INCLUDE_ALL_METADATA")
    metadata_fields      = try(var.spec.log_config.metadata_fields, [])
    filter_expr          = var.spec.log_config.filter_expr != "" ? var.spec.log_config.filter_expr : null
  }
}
