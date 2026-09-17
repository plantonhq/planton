locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain; an empty string would be sent
  # verbatim and rejected by the API.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # api_key_id is Optional+Computed on the provider: when the spec names no
  # key, Firebase associates or provisions one and the computed value is read
  # back, so null (not "") is sent when the spec leaves it empty.
  api_key_id = var.spec.api_key_id != "" ? var.spec.api_key_id : null

  # app_store_id and team_id are plain Optional on the provider (not
  # Computed): sent only when set, so an unset field never diffs.
  app_store_id = var.spec.app_store_id != "" ? var.spec.app_store_id : null
  team_id      = var.spec.team_id != "" ? var.spec.team_id : null

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # App Check. App Attest counts as configured when its block is present and
  # not declared off (enabled defaults to true -- the wire has no switch, the
  # configuration's existence is the switch). DeviceCheck's block carries
  # required content, so its presence is its switch.
  wants_app_attest   = var.spec.app_check != null && var.spec.app_check.app_attest != null && (var.spec.app_check.app_attest.enabled == null || var.spec.app_check.app_attest.enabled)
  wants_device_check = var.spec.app_check != null && var.spec.app_check.device_check != null
  debug_tokens       = var.spec.app_check != null ? { for t in var.spec.app_check.debug_tokens : t.display_name => t } : {}
  wants_app_check    = local.wants_app_attest || local.wants_device_check || length(local.debug_tokens) > 0
}
