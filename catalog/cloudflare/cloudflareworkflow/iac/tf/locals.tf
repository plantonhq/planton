locals {
  # Retention values are dynamic at the API (integer milliseconds or a
  # duration expression like "5 minutes"); the spec carries both forms as
  # strings and they pass through verbatim. Empty strings mean "not set" --
  # drop them so Cloudflare keeps its defaults. The whole block goes null
  # when neither side is set.
  error_retention   = try(var.spec.default_retention.error_retention, "") != "" ? var.spec.default_retention.error_retention : null
  success_retention = try(var.spec.default_retention.success_retention, "") != "" ? var.spec.default_retention.success_retention : null
  default_retention = (local.error_retention != null || local.success_retention != null) ? {
    error_retention   = local.error_retention
    success_retention = local.success_retention
  } : null

  # An absent limits block keeps Cloudflare's defaults.
  limits = try(var.spec.limits.steps, null) != null ? { steps = var.spec.limits.steps } : null

  schedules = length(coalesce(var.spec.schedules, [])) > 0 ? [
    for schedule in var.spec.schedules : { cron = schedule.cron }
  ] : null
}
