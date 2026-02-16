resource "google_kms_crypto_key" "this" {
  name     = local.key_name
  key_ring = local.key_ring_id

  purpose                       = local.purpose
  rotation_period               = local.rotation_period
  destroy_scheduled_duration    = local.destroy_scheduled_duration
  skip_initial_version_creation = local.skip_initial_version_creation

  dynamic "version_template" {
    for_each = local.version_template != null ? [local.version_template] : []
    content {
      algorithm        = version_template.value.algorithm
      protection_level = version_template.value.protection_level != "" ? version_template.value.protection_level : null
    }
  }
}
