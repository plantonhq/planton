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

  # Cloudflare stores a name-valued record's content without the root dot
  # (`ns-cloud-a1.googledomains.com.` is kept as `ns-cloud-a1.googledomains.com`),
  # and this provider compares the stored value to the configured one verbatim,
  # so a content that ends in a dot would plan an update on every run. Cloud DNS
  # and most DNS tooling spell names absolutely, and a delegation's content is
  # often another zone's nameserver output, so the dot is dropped here rather
  # than asked of every author. The Pulumi provider reconciles the dot itself.
  name_valued = contains(["CNAME", "NS", "MX", "PTR"], local.record_type)
  content     = var.spec.content == "" ? null : (local.name_valued ? trimsuffix(var.spec.content, ".") : var.spec.content)

  # The provider schema marks top-level priority "Required for MX, SRV and URI
  # records". MX carries it in spec.priority; SRV/URI carry it inside their
  # structured data, and Cloudflare mirrors that value into the top-level field
  # on its own -- leaving config null there drifts forever (live-measured: an
  # SRV record's re-plan wanted `priority = 10 -> null`). Mirror the API.
  top_priority = (
    local.record_type == "MX" ? var.spec.priority :
    local.record_type == "SRV" ? try(var.spec.srv.priority, null) :
    local.record_type == "URI" ? try(var.spec.uri.priority, null) :
    null
  )

  # The proto models structured data as a oneof, and the tfvars converter emits
  # the ACTIVE CASE as a top-level spec attribute (srv = {...}, caa = {...}); a
  # "data" wrapper key never appears in tfvars (the zone kind's records.tf is
  # the same idiom).
  record_has_data = (
    var.spec.caa != null || var.spec.cert != null || var.spec.dnskey != null ||
    var.spec.ds != null || var.spec.https != null || var.spec.loc != null ||
    var.spec.naptr != null || var.spec.smimea != null || var.spec.srv != null ||
    var.spec.sshfp != null || var.spec.svcb != null || var.spec.tlsa != null ||
    var.spec.uri != null
  )

  # Flatten the active structured case into the provider's single flat data
  # object. At most one case is populated; for each provider data attribute,
  # try() walks the cases that carry it (an unset case errors on attribute
  # access and is skipped, while the active case's omitted optionals return
  # null).
  record_data = !local.record_has_data ? null : {
    flags          = try(var.spec.caa.flags, var.spec.dnskey.flags, var.spec.naptr.flags, null)
    tag            = try(var.spec.caa.tag, null)
    value          = try(var.spec.caa.value, var.spec.https.value, var.spec.svcb.value, null)
    type           = try(var.spec.cert.type, var.spec.sshfp.type, null)
    key_tag        = try(var.spec.cert.key_tag, var.spec.ds.key_tag, null)
    algorithm      = try(var.spec.cert.algorithm, var.spec.dnskey.algorithm, var.spec.ds.algorithm, var.spec.sshfp.algorithm, null)
    certificate    = try(var.spec.cert.certificate, var.spec.smimea.certificate, var.spec.tlsa.certificate, null)
    protocol       = try(var.spec.dnskey.protocol, null)
    public_key     = try(var.spec.dnskey.public_key, null)
    digest         = try(var.spec.ds.digest, null)
    digest_type    = try(var.spec.ds.digest_type, null)
    priority       = try(var.spec.https.priority, var.spec.srv.priority, var.spec.svcb.priority, var.spec.uri.priority, null)
    target         = try(var.spec.https.target, var.spec.srv.target, var.spec.svcb.target, var.spec.uri.target, null)
    altitude       = try(var.spec.loc.altitude, null)
    lat_degrees    = try(var.spec.loc.lat_degrees, null)
    lat_direction  = try(var.spec.loc.lat_direction, null)
    lat_minutes    = try(var.spec.loc.lat_minutes, null)
    lat_seconds    = try(var.spec.loc.lat_seconds, null)
    long_degrees   = try(var.spec.loc.long_degrees, null)
    long_direction = try(var.spec.loc.long_direction, null)
    long_minutes   = try(var.spec.loc.long_minutes, null)
    long_seconds   = try(var.spec.loc.long_seconds, null)
    precision_horz = try(var.spec.loc.precision_horz, null)
    precision_vert = try(var.spec.loc.precision_vert, null)
    size           = try(var.spec.loc.size, null)
    order          = try(var.spec.naptr.order, null)
    preference     = try(var.spec.naptr.preference, null)
    regex          = try(var.spec.naptr.regex, null)
    replacement    = try(var.spec.naptr.replacement, null)
    service        = try(var.spec.naptr.service, null)
    matching_type  = try(var.spec.smimea.matching_type, var.spec.tlsa.matching_type, null)
    selector       = try(var.spec.smimea.selector, var.spec.tlsa.selector, null)
    usage          = try(var.spec.smimea.usage, var.spec.tlsa.usage, null)
    port           = try(var.spec.srv.port, null)
    weight         = try(var.spec.srv.weight, var.spec.uri.weight, null)
    fingerprint    = try(var.spec.sshfp.fingerprint, null)
  }
}
