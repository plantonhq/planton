output "certificate_id" {
  description = "The Hetzner Cloud numeric ID of the created certificate"
  value = (
    local.is_uploaded
    ? hcloud_uploaded_certificate.this[0].id
    : hcloud_managed_certificate.this[0].id
  )
}

output "type" {
  description = "Certificate type: uploaded or managed"
  value = (
    local.is_uploaded
    ? hcloud_uploaded_certificate.this[0].type
    : hcloud_managed_certificate.this[0].type
  )
}

output "fingerprint" {
  description = "SHA256 fingerprint of the certificate"
  value = (
    local.is_uploaded
    ? hcloud_uploaded_certificate.this[0].fingerprint
    : hcloud_managed_certificate.this[0].fingerprint
  )
}

output "not_valid_before" {
  description = "Point in time when the certificate becomes valid (ISO-8601)"
  value = (
    local.is_uploaded
    ? hcloud_uploaded_certificate.this[0].not_valid_before
    : hcloud_managed_certificate.this[0].not_valid_before
  )
}

output "not_valid_after" {
  description = "Point in time when the certificate stops being valid (ISO-8601)"
  value = (
    local.is_uploaded
    ? hcloud_uploaded_certificate.this[0].not_valid_after
    : hcloud_managed_certificate.this[0].not_valid_after
  )
}
