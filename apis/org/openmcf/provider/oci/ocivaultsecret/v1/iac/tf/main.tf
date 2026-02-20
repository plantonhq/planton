resource "oci_vault_secret" "this" {
  compartment_id = var.spec.compartment_id.value
  secret_name    = var.spec.secret_name
  vault_id       = var.spec.vault_id.value
  key_id         = var.spec.key_id.value
  freeform_tags  = local.freeform_tags

  description          = var.spec.description != "" ? var.spec.description : null
  enable_auto_generation = var.spec.enable_auto_generation ? true : null
  metadata             = length(var.spec.secret_metadata) > 0 ? var.spec.secret_metadata : null

  dynamic "secret_content" {
    for_each = var.spec.secret_content != null ? [var.spec.secret_content] : []
    content {
      content_type = "BASE64"
      content      = secret_content.value.content != "" ? secret_content.value.content : null
      name         = secret_content.value.name != "" ? secret_content.value.name : null
      stage        = secret_content.value.stage != "" ? secret_content.value.stage : null
    }
  }

  dynamic "secret_generation_context" {
    for_each = var.spec.secret_generation_context != null ? [var.spec.secret_generation_context] : []
    content {
      generation_type     = lookup(local.generation_type_map, secret_generation_context.value.generation_type, secret_generation_context.value.generation_type)
      generation_template = secret_generation_context.value.generation_template
      passphrase_length   = secret_generation_context.value.passphrase_length > 0 ? secret_generation_context.value.passphrase_length : null
      secret_template     = secret_generation_context.value.secret_template != "" ? secret_generation_context.value.secret_template : null
    }
  }

  dynamic "secret_rules" {
    for_each = var.spec.secret_rules
    content {
      rule_type                                        = lookup(local.rule_type_map, secret_rules.value.rule_type, secret_rules.value.rule_type)
      is_secret_content_retrieval_blocked_on_expiry     = secret_rules.value.is_secret_content_retrieval_blocked_on_expiry ? true : null
      secret_version_expiry_interval                    = secret_rules.value.secret_version_expiry_interval != "" ? secret_rules.value.secret_version_expiry_interval : null
      time_of_absolute_expiry                           = secret_rules.value.time_of_absolute_expiry != "" ? secret_rules.value.time_of_absolute_expiry : null
      is_enforced_on_deleted_secret_versions            = secret_rules.value.is_enforced_on_deleted_secret_versions ? true : null
    }
  }

  dynamic "rotation_config" {
    for_each = var.spec.rotation_config != null ? [var.spec.rotation_config] : []
    content {
      is_scheduled_rotation_enabled = rotation_config.value.is_scheduled_rotation_enabled ? true : null
      rotation_interval             = rotation_config.value.rotation_interval != "" ? rotation_config.value.rotation_interval : null

      target_system_details {
        target_system_type = lookup(local.target_system_type_map, rotation_config.value.target_system_details.target_system_type, rotation_config.value.target_system_details.target_system_type)
        adb_id             = rotation_config.value.target_system_details.adb_id != null ? rotation_config.value.target_system_details.adb_id.value : null
        function_id        = rotation_config.value.target_system_details.function_id != null ? rotation_config.value.target_system_details.function_id.value : null
      }
    }
  }
}
