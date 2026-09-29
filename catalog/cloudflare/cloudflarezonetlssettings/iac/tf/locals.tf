locals {
  zone_id = var.spec.zone_id

  # Per-hostname overrides, split per setting_id: each set attribute of each
  # row is its own API object keyed by (setting_id, hostname). The hostname is
  # the for_each key so editing one row never churns another row's resources.
  hostname_min_tls_version = {
    for row in var.spec.hostname_settings : row.hostname => row.min_tls_version
    if row.min_tls_version != null
  }
  hostname_http2 = {
    for row in var.spec.hostname_settings : row.hostname => row.http2
    if row.http2 != null
  }
  hostname_ciphers = {
    for row in var.spec.hostname_settings : row.hostname => row.ciphers
    if length(row.ciphers) > 0
  }

  # CA hostname associations, keyed by the mTLS certificate id (or "managed"
  # for the zone's managed-CA list -- the row without a certificate id).
  ca_hostname_associations = {
    for association in var.spec.ca_hostname_associations :
    (association.mtls_certificate_id == "" ? "managed" : association.mtls_certificate_id) => association
  }
}
