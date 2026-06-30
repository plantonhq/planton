resource "alicloud_kms_key" "main" {
  description                      = var.spec.description != "" ? var.spec.description : null
  key_spec                         = var.spec.key_spec
  key_usage                        = var.spec.key_usage
  protection_level                 = var.spec.protection_level
  automatic_rotation               = local.automatic_rotation
  rotation_interval                = var.spec.rotation_interval != "" ? var.spec.rotation_interval : null
  pending_window_in_days           = var.spec.pending_window_in_days
  deletion_protection              = local.deletion_protection
  deletion_protection_description  = var.spec.deletion_protection_description != "" ? var.spec.deletion_protection_description : null
  tags                             = local.final_tags
}
