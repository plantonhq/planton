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

  # The cloud-side name defaults to metadata.name when the spec leaves
  # backend_bucket_name empty — the same naming basis every kind uses.
  backend_bucket_name = (
    var.spec.backend_bucket_name != null && var.spec.backend_bucket_name != ""
    ? var.spec.backend_bucket_name
    : var.metadata.name
  )

  # Normalize "" -> null for optional strings the provider treats as
  # meaningfully absent (an empty compression_mode or scheme would be
  # rejected; an empty edge policy would try to clear one that was never
  # set).
  compression_mode      = var.spec.compression_mode != "" ? var.spec.compression_mode : null
  load_balancing_scheme = var.spec.load_balancing_scheme != "" ? var.spec.load_balancing_scheme : null
  edge_security_policy  = var.spec.edge_security_policy != "" ? var.spec.edge_security_policy : null

  # The tfvars converter emits 0 for unset proto ints, but a real 0 TTL is
  # also meaningful to GCP ("don't cache"). The spec resolves the ambiguity
  # in favor of the overwhelmingly common intent: 0 = unset, letting the API
  # apply its own defaults. cache_mode governs which TTLs may be sent at all
  # (GCP rejects TTLs it would ignore) — the spec's CEL rules enforce that
  # before deploy, so no TTL-stripping logic is needed here.
  # request_coalescing is sent only when enabled: the API defaults coalescing
  # ON, so an explicit false would switch it off for a manifest that never
  # mentioned it.
  cdn_policy = var.spec.cdn_policy == null ? null : {
    cache_mode                      = try(var.spec.cdn_policy.cache_mode, null) != "" ? var.spec.cdn_policy.cache_mode : null
    client_ttl                      = try(var.spec.cdn_policy.client_ttl, 0) != 0 ? var.spec.cdn_policy.client_ttl : null
    default_ttl                     = try(var.spec.cdn_policy.default_ttl, 0) != 0 ? var.spec.cdn_policy.default_ttl : null
    max_ttl                         = try(var.spec.cdn_policy.max_ttl, 0) != 0 ? var.spec.cdn_policy.max_ttl : null
    negative_caching                = try(var.spec.cdn_policy.negative_caching, null)
    negative_caching_policy         = try(var.spec.cdn_policy.negative_caching_policy, [])
    serve_while_stale               = try(var.spec.cdn_policy.serve_while_stale, 0) != 0 ? var.spec.cdn_policy.serve_while_stale : null
    request_coalescing              = try(var.spec.cdn_policy.request_coalescing, false) ? true : null
    signed_url_cache_max_age_sec    = try(var.spec.cdn_policy.signed_url_cache_max_age_sec, 0) != 0 ? var.spec.cdn_policy.signed_url_cache_max_age_sec : null
    cache_key_policy                = try(var.spec.cdn_policy.cache_key_policy, null)
    bypass_cache_on_request_headers = try(var.spec.cdn_policy.bypass_cache_on_request_headers, [])
  }
}
