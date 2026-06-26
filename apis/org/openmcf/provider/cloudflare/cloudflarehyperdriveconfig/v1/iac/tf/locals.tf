locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-hyperdrive-config")

  origin = var.spec.origin

  # Caching is set only when the block is provided; within it, 0 means "use the
  # provider default", so those fields are passed as null.
  caching = var.spec.caching != null ? {
    disabled               = try(var.spec.caching.disabled, false)
    max_age                = try(var.spec.caching.max_age, 0) > 0 ? var.spec.caching.max_age : null
    stale_while_revalidate = try(var.spec.caching.stale_while_revalidate, 0) > 0 ? var.spec.caching.stale_while_revalidate : null
  } : null

  # mTLS is set only when at least one field is provided.
  mtls_provided = var.spec.mtls != null && (
    try(var.spec.mtls.ca_certificate_id, "") != "" ||
    try(var.spec.mtls.mtls_certificate_id, "") != "" ||
    try(var.spec.mtls.sslmode, "") != ""
  )
  mtls = local.mtls_provided ? {
    ca_certificate_id   = try(var.spec.mtls.ca_certificate_id, "") != "" ? var.spec.mtls.ca_certificate_id : null
    mtls_certificate_id = try(var.spec.mtls.mtls_certificate_id, "") != "" ? var.spec.mtls.mtls_certificate_id : null
    sslmode             = try(var.spec.mtls.sslmode, "") != "" ? var.spec.mtls.sslmode : null
  } : null
}
