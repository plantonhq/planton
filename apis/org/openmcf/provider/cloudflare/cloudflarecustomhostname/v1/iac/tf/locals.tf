locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-custom-hostname")

  # Zone + origin (StringValueOrRef flattened to plain strings by the converter).
  zone_id              = try(var.spec.zone_id, "")
  custom_origin_server = try(var.spec.custom_origin_server, "") != "" ? var.spec.custom_origin_server : null
  custom_origin_sni    = try(var.spec.custom_origin_sni, "") != "" ? var.spec.custom_origin_sni : null
  custom_metadata      = length(try(var.spec.custom_metadata, {})) > 0 ? var.spec.custom_metadata : null

  # Build the ssl object with defaults coalesced (bundle_method "ubiquitous", type
  # "dv") and unset optionals omitted (null) so behavior matches the Pulumi module
  # byte-for-byte. Only configurable attributes are set; computed sub-attributes
  # (validation records/errors, status) are left to the provider.
  ssl = try(var.spec.ssl, null) == null ? null : {
    bundle_method         = try(var.spec.ssl.bundle_method, "") != "" ? var.spec.ssl.bundle_method : "ubiquitous"
    type                  = try(var.spec.ssl.type, "") != "" ? var.spec.ssl.type : "dv"
    certificate_authority = try(var.spec.ssl.certificate_authority, "") != "" ? var.spec.ssl.certificate_authority : null
    cloudflare_branding   = try(var.spec.ssl.cloudflare_branding, false) ? true : null
    method                = try(var.spec.ssl.method, "") != "" ? var.spec.ssl.method : null
    wildcard              = try(var.spec.ssl.wildcard, false) ? true : null
    custom_certificate    = try(var.spec.ssl.custom_certificate, "") != "" ? var.spec.ssl.custom_certificate : null
    custom_csr_id         = try(var.spec.ssl.custom_csr_id, "") != "" ? var.spec.ssl.custom_csr_id : null
    custom_key            = try(var.spec.ssl.custom_key, "") != "" ? var.spec.ssl.custom_key : null
    custom_cert_bundle    = length(try(var.spec.ssl.custom_cert_bundle, [])) > 0 ? var.spec.ssl.custom_cert_bundle : null
    settings = try(var.spec.ssl.settings, null) == null ? null : {
      ciphers         = length(try(var.spec.ssl.settings.ciphers, [])) > 0 ? var.spec.ssl.settings.ciphers : null
      early_hints     = try(var.spec.ssl.settings.early_hints, "") != "" ? var.spec.ssl.settings.early_hints : null
      http2           = try(var.spec.ssl.settings.http2, "") != "" ? var.spec.ssl.settings.http2 : null
      min_tls_version = try(var.spec.ssl.settings.min_tls_version, "") != "" ? var.spec.ssl.settings.min_tls_version : null
      tls_1_3         = try(var.spec.ssl.settings.tls_1_3, "") != "" ? var.spec.ssl.settings.tls_1_3 : null
    }
  }
}
