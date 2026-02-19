resource "alicloud_nas_file_system" "main" {
  file_system_type  = var.spec.file_system_type
  protocol_type     = var.spec.protocol_type
  storage_type      = var.spec.storage_type
  description       = var.spec.description != "" ? var.spec.description : null
  resource_group_id = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags              = local.final_tags

  encrypt_type = var.spec.encryption != null ? var.spec.encryption.encrypt_type : null
  kms_key_id   = var.spec.encryption != null && var.spec.encryption.kms_key_id != "" ? var.spec.encryption.kms_key_id : null

  capacity = var.spec.capacity > 0 ? var.spec.capacity : null
  zone_id  = var.spec.zone_id != "" ? var.spec.zone_id : null

  vpc_id     = var.spec.file_system_type == "extreme" ? var.spec.vpc_id : null
  vswitch_id = var.spec.file_system_type == "extreme" ? var.spec.vswitch_id : null
}

resource "alicloud_nas_mount_target" "main" {
  file_system_id   = alicloud_nas_file_system.main.id
  access_group_name = local.has_access_rules ? alicloud_nas_access_group.custom[0].access_group_name : null
  vpc_id           = var.spec.vpc_id
  vswitch_id       = var.spec.vswitch_id
}
