resource "alicloud_instance" "main" {
  instance_type      = var.spec.instance_type
  image_id           = var.spec.image_id
  vswitch_id         = var.spec.vswitch_id
  security_groups    = var.spec.security_group_ids
  instance_name      = local.instance_name
  instance_charge_type = var.spec.instance_charge_type

  host_name   = var.spec.host_name != "" ? var.spec.host_name : null
  description = var.spec.description != "" ? var.spec.description : null

  system_disk_category          = var.spec.system_disk.category
  system_disk_size              = var.spec.system_disk.size
  system_disk_performance_level = var.spec.system_disk.performance_level != "" ? var.spec.system_disk.performance_level : null
  system_disk_encrypted         = var.spec.system_disk.encrypted
  system_disk_kms_key_id        = var.spec.system_disk.kms_key_id != "" ? var.spec.system_disk.kms_key_id : null

  dynamic "data_disks" {
    for_each = var.spec.data_disks
    content {
      size                 = data_disks.value.size
      category             = data_disks.value.category
      name                 = data_disks.value.name != "" ? data_disks.value.name : null
      performance_level    = data_disks.value.performance_level != "" ? data_disks.value.performance_level : null
      encrypted            = data_disks.value.encrypted
      kms_key_id           = data_disks.value.kms_key_id != "" ? data_disks.value.kms_key_id : null
      snapshot_id          = data_disks.value.snapshot_id != "" ? data_disks.value.snapshot_id : null
      delete_with_instance = data_disks.value.delete_with_instance
      description          = data_disks.value.description != "" ? data_disks.value.description : null
    }
  }

  key_name = var.spec.key_name != "" ? var.spec.key_name : null
  password = var.spec.password != "" ? var.spec.password : null

  internet_max_bandwidth_out = var.spec.internet_max_bandwidth_out
  internet_charge_type       = var.spec.internet_charge_type != "" ? var.spec.internet_charge_type : null
  period                     = var.spec.period
  period_unit                = var.spec.period_unit != "" ? var.spec.period_unit : null

  spot_strategy    = var.spec.spot_strategy != "" ? var.spec.spot_strategy : null
  spot_price_limit = var.spec.spot_price_limit

  user_data  = var.spec.user_data != "" ? var.spec.user_data : null
  role_name  = var.spec.role_name != "" ? var.spec.role_name : null

  deletion_protection             = var.spec.deletion_protection
  security_enhancement_strategy   = var.spec.security_enhancement_strategy != "" ? var.spec.security_enhancement_strategy : null
  resource_group_id               = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null

  tags = local.final_tags
}
