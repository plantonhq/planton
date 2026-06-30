# locals.tf

locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "cloudflare_dns_record"
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null &&
    try(var.metadata.env, "") != ""
  ) ? { "environment" = var.metadata.env } : {}

  # Merge base, org, and environment labels
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Normalize record type to uppercase
  record_type = upper(var.spec.type)

  # Only A, AAAA, and CNAME records may be proxied
  supports_proxy = contains(["A", "AAAA", "CNAME"], local.record_type)
  proxied        = local.supports_proxy ? var.spec.proxied : false

  # Priority is only used for MX records
  requires_priority = local.record_type == "MX"

  # Flatten the structured `data` oneof into the provider's single flat data
  # object. Exactly one case is populated; for each provider data attribute, try()
  # walks the cases that carry it (an unset case errors on attribute access and is
  # skipped, while the active case's omitted optionals return null).
  record_data = var.spec.data == null ? null : {
    flags         = try(var.spec.data.caa.flags, var.spec.data.dnskey.flags, var.spec.data.naptr.flags, null)
    tag           = try(var.spec.data.caa.tag, null)
    value         = try(var.spec.data.caa.value, var.spec.data.https.value, var.spec.data.svcb.value, null)
    type          = try(var.spec.data.cert.type, var.spec.data.sshfp.type, null)
    key_tag       = try(var.spec.data.cert.key_tag, var.spec.data.ds.key_tag, null)
    algorithm     = try(var.spec.data.cert.algorithm, var.spec.data.dnskey.algorithm, var.spec.data.ds.algorithm, var.spec.data.sshfp.algorithm, null)
    certificate   = try(var.spec.data.cert.certificate, var.spec.data.smimea.certificate, var.spec.data.tlsa.certificate, null)
    protocol      = try(var.spec.data.dnskey.protocol, null)
    public_key    = try(var.spec.data.dnskey.public_key, null)
    digest        = try(var.spec.data.ds.digest, null)
    digest_type   = try(var.spec.data.ds.digest_type, null)
    priority      = try(var.spec.data.https.priority, var.spec.data.srv.priority, var.spec.data.svcb.priority, var.spec.data.uri.priority, null)
    target        = try(var.spec.data.https.target, var.spec.data.srv.target, var.spec.data.svcb.target, var.spec.data.uri.target, null)
    altitude      = try(var.spec.data.loc.altitude, null)
    lat_degrees   = try(var.spec.data.loc.lat_degrees, null)
    lat_direction = try(var.spec.data.loc.lat_direction, null)
    lat_minutes   = try(var.spec.data.loc.lat_minutes, null)
    lat_seconds   = try(var.spec.data.loc.lat_seconds, null)
    long_degrees  = try(var.spec.data.loc.long_degrees, null)
    long_direction = try(var.spec.data.loc.long_direction, null)
    long_minutes  = try(var.spec.data.loc.long_minutes, null)
    long_seconds  = try(var.spec.data.loc.long_seconds, null)
    precision_horz = try(var.spec.data.loc.precision_horz, null)
    precision_vert = try(var.spec.data.loc.precision_vert, null)
    size          = try(var.spec.data.loc.size, null)
    order         = try(var.spec.data.naptr.order, null)
    preference    = try(var.spec.data.naptr.preference, null)
    regex         = try(var.spec.data.naptr.regex, null)
    replacement   = try(var.spec.data.naptr.replacement, null)
    service       = try(var.spec.data.naptr.service, null)
    matching_type = try(var.spec.data.smimea.matching_type, var.spec.data.tlsa.matching_type, null)
    selector      = try(var.spec.data.smimea.selector, var.spec.data.tlsa.selector, null)
    usage         = try(var.spec.data.smimea.usage, var.spec.data.tlsa.usage, null)
    port          = try(var.spec.data.srv.port, null)
    weight        = try(var.spec.data.srv.weight, var.spec.data.uri.weight, null)
    fingerprint   = try(var.spec.data.sshfp.fingerprint, null)
  }
}
