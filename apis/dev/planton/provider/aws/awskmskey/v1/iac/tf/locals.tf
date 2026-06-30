locals {
  resource_name = coalesce(try(var.metadata.name, null), "aws-kms-key")
  tags          = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  key_spec_normalized      = upper(try(var.spec.key_spec, "SYMMETRIC"))
  customer_master_key_spec = local.key_spec_normalized == "SYMMETRIC" ? "SYMMETRIC_DEFAULT" : local.key_spec_normalized

  is_rotation_disabled = try(var.spec.disable_key_rotation, false)
  deletion_window_days = try(var.spec.deletion_window_days, 30)
  description          = try(var.spec.description, null)
  alias_name           = try(var.spec.alias_name, null)

  # Rotation only applicable for symmetric keys
  rotation_enabled = (local.key_spec_normalized == "SYMMETRIC") && (!local.is_rotation_disabled)
}



