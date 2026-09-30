# records.tf

locals {
  # Key includes the list index so duplicate name+type pairs (legal in DNS,
  # e.g. round-robin A records) stay unique.
  record_entries = { for idx, record in var.spec.records : "${record.name}-${record.type}-${idx}" => record }

  # The proto models structured data as a oneof, and the tfvars converter emits
  # the ACTIVE CASE as a top-level attribute on the record (srv = {...}, caa =
  # {...}); a "data" wrapper key never appears in tfvars.
  record_has_data = {
    for key, record in local.record_entries : key => (
      record.caa != null || record.cert != null || record.dnskey != null ||
      record.ds != null || record.https != null || record.loc != null ||
      record.naptr != null || record.smimea != null || record.srv != null ||
      record.sshfp != null || record.svcb != null || record.tlsa != null ||
      record.uri != null
    )
  }

  # Flatten each record's structured case into the provider's single flat data
  # object. Exactly one case is populated; for each provider data attribute,
  # try() walks the cases that carry it (an unset case errors on attribute
  # access and is skipped, while the active case's omitted optionals return
  # null).
  record_data = {
    for key, record in local.record_entries : key => (
      !local.record_has_data[key] ? null : {
        flags          = try(record.caa.flags, record.dnskey.flags, record.naptr.flags, null)
        tag            = try(record.caa.tag, null)
        value          = try(record.caa.value, record.https.value, record.svcb.value, null)
        type           = try(record.cert.type, record.sshfp.type, null)
        key_tag        = try(record.cert.key_tag, record.ds.key_tag, null)
        algorithm      = try(record.cert.algorithm, record.dnskey.algorithm, record.ds.algorithm, record.sshfp.algorithm, null)
        certificate    = try(record.cert.certificate, record.smimea.certificate, record.tlsa.certificate, null)
        protocol       = try(record.dnskey.protocol, null)
        public_key     = try(record.dnskey.public_key, null)
        digest         = try(record.ds.digest, null)
        digest_type    = try(record.ds.digest_type, null)
        priority       = try(record.https.priority, record.srv.priority, record.svcb.priority, record.uri.priority, null)
        target         = try(record.https.target, record.srv.target, record.svcb.target, record.uri.target, null)
        altitude       = try(record.loc.altitude, null)
        lat_degrees    = try(record.loc.lat_degrees, null)
        lat_direction  = try(record.loc.lat_direction, null)
        lat_minutes    = try(record.loc.lat_minutes, null)
        lat_seconds    = try(record.loc.lat_seconds, null)
        long_degrees   = try(record.loc.long_degrees, null)
        long_direction = try(record.loc.long_direction, null)
        long_minutes   = try(record.loc.long_minutes, null)
        long_seconds   = try(record.loc.long_seconds, null)
        precision_horz = try(record.loc.precision_horz, null)
        precision_vert = try(record.loc.precision_vert, null)
        size           = try(record.loc.size, null)
        order          = try(record.naptr.order, null)
        preference     = try(record.naptr.preference, null)
        regex          = try(record.naptr.regex, null)
        replacement    = try(record.naptr.replacement, null)
        service        = try(record.naptr.service, null)
        matching_type  = try(record.smimea.matching_type, record.tlsa.matching_type, null)
        selector       = try(record.smimea.selector, record.tlsa.selector, null)
        usage          = try(record.smimea.usage, record.tlsa.usage, null)
        port           = try(record.srv.port, null)
        weight         = try(record.srv.weight, record.uri.weight, null)
        fingerprint    = try(record.sshfp.fingerprint, null)
      }
    )
  }
}

# Create DNS records within the zone
resource "cloudflare_dns_record" "records" {
  for_each = local.record_entries

  zone_id = cloudflare_zone.main.id
  name    = each.value.name
  type    = each.value.type

  # The provider requires ttl >= 1 (1 = automatic); the proto's unset 0 maps to
  # automatic.
  ttl = each.value.ttl > 0 ? each.value.ttl : 1

  # Simple record types carry their value in content; structured types use data.
  # Cloudflare stores a name-valued record's content without the root dot, and
  # this provider compares the stored value to the configured one verbatim, so
  # an NS, CNAME, MX or PTR content ending in a dot (the everyday spelling of
  # another zone's nameserver output) would plan an update on every run. The
  # dot is dropped here, as the record kind's module does; the Pulumi provider
  # reconciles it itself.
  content = each.value.content == "" ? null : (
    contains(["CNAME", "NS", "MX", "PTR"], upper(each.value.type)) ? trimsuffix(each.value.content, ".") : each.value.content
  )
  data    = local.record_data[each.key]

  # proxied is only applicable to A, AAAA, and CNAME records
  proxied = contains(["A", "AAAA", "CNAME"], each.value.type) ? each.value.proxied : false

  # The provider schema marks top-level priority "Required for MX, SRV and URI
  # records". MX carries it in spec priority; SRV/URI carry it inside their
  # structured data, and Cloudflare mirrors that value into the top-level
  # field on its own -- omitting the mirror here drifts on refresh forever
  # (live-measured on the sibling record kind's SRV lane). HTTPS/SVCB carry
  # priority only inside data.
  priority = (
    each.value.type == "MX" ? each.value.priority :
    each.value.type == "SRV" ? try(each.value.srv.priority, null) :
    each.value.type == "URI" ? try(each.value.uri.priority, null) :
    null
  )

  # comment for the DNS record
  comment = each.value.comment != "" ? each.value.comment : null

  # Custom tags
  tags = length(each.value.tags) > 0 ? toset(each.value.tags) : null

  # Record-level settings (only affect proxied records)
  settings = each.value.settings

  # Restrict the record to Cloudflare internal (private) routing when set.
  private_routing = each.value.private_routing ? true : null

  depends_on = [cloudflare_zone.main]
}
