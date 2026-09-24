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

  # Certificate fingerprints are sent only when declared: the provider's
  # lists are Optional and an absent list means "none registered", which is
  # exactly what an empty spec list means too.
  sha1_hashes   = length(var.spec.sha1_hashes) > 0 ? var.spec.sha1_hashes : null
  sha256_hashes = length(var.spec.sha256_hashes) > 0 ? var.spec.sha256_hashes : null

  # App Check. Play Integrity counts as configured when its block is present
  # and not declared off (enabled defaults to true -- the wire has no switch,
  # the configuration's existence is the switch).
  wants_play_integrity = var.spec.app_check != null && var.spec.app_check.play_integrity != null && (var.spec.app_check.play_integrity.enabled == null || var.spec.app_check.play_integrity.enabled)
  debug_tokens         = var.spec.app_check != null ? { for t in var.spec.app_check.debug_tokens : t.display_name => t } : {}
  wants_app_check      = local.wants_play_integrity || length(local.debug_tokens) > 0
}
