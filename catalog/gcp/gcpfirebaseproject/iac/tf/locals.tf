locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain; an empty string would be sent
  # verbatim and rejected by the API.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The spec's deletion_policy, sent only when set so the provider's default
  # (DELETE) stays the provider's. Governs the composed resources only; the
  # enablement itself is undeletable.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Whether the spec composes the optional surfaces -- gates the API
  # enablement and the resources together so a spec without App Check
  # enables nothing App-Check-shaped.
  wants_default_bucket = var.spec.default_storage_location != ""
  wants_app_check      = var.spec.app_check != null && (length(var.spec.app_check.service_configs) > 0 || length(var.spec.app_check.resource_policies) > 0)

  # The App Check configurations keyed for for_each: service configs by
  # their immutable service id; resource policies by service id plus target
  # resource -- both immutable identity keys, so plans stay stable as list
  # order changes.
  app_check_service_configs = local.wants_app_check ? {
    for sc in var.spec.app_check.service_configs : sc.service_id => sc
  } : {}
  app_check_resource_policies = local.wants_app_check ? {
    for rp in var.spec.app_check.resource_policies : "${rp.service_id}/${rp.target_resource}" => rp
  } : {}
}
