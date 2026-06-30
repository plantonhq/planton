resource "oci_file_storage_export" "this" {
  for_each = { for e in var.spec.exports : e.path => e }

  export_set_id  = oci_file_storage_mount_target.this.export_set_id
  file_system_id = oci_file_storage_file_system.this.id
  path           = each.value.path

  dynamic "export_options" {
    for_each = each.value.export_options
    content {
      source                         = export_options.value.source
      access                         = export_options.value.access != "" ? lookup(local.access_map, export_options.value.access, null) : null
      identity_squash                = export_options.value.identity_squash != "" ? lookup(local.identity_squash_map, export_options.value.identity_squash, null) : null
      require_privileged_source_port = export_options.value.require_privileged_source_port
      is_anonymous_access_allowed    = export_options.value.is_anonymous_access_allowed
      anonymous_uid                  = export_options.value.anonymous_uid != 0 ? tostring(export_options.value.anonymous_uid) : null
      anonymous_gid                  = export_options.value.anonymous_gid != 0 ? tostring(export_options.value.anonymous_gid) : null
    }
  }
}
