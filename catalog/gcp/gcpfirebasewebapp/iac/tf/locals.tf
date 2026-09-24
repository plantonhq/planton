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

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # App Check. Both reCAPTCHA blocks carry required content, so their
  # presence is their switch; both may be set at once (the v3-to-Enterprise
  # migration shape).
  wants_recaptcha_v3         = var.spec.app_check != null && var.spec.app_check.recaptcha_v3 != null
  wants_recaptcha_enterprise = var.spec.app_check != null && var.spec.app_check.recaptcha_enterprise != null
  debug_tokens               = var.spec.app_check != null ? { for t in var.spec.app_check.debug_tokens : t.display_name => t } : {}
  wants_app_check            = local.wants_recaptcha_v3 || local.wants_recaptcha_enterprise || length(local.debug_tokens) > 0
}
