resource "alicloud_cs_kubernetes_node_pool" "node_pool" {
  cluster_id     = var.spec.cluster_id
  node_pool_name = local.node_pool_name
  vswitch_ids    = var.spec.vswitch_ids
  instance_types = var.spec.instance_types
  desired_size   = var.spec.desired_size
  image_type     = var.spec.image_type
  tags           = local.final_tags

  system_disk_category          = local.system_disk.category
  system_disk_size              = local.system_disk.size
  system_disk_performance_level = local.system_disk.performance_level != "" ? local.system_disk.performance_level : null
  system_disk_encrypted         = local.system_disk.encrypted
  system_disk_kms_key           = local.system_disk.kms_key_id != "" ? local.system_disk.kms_key_id : null

  dynamic "data_disks" {
    for_each = var.spec.data_disks
    content {
      category          = data_disks.value.category
      size              = data_disks.value.size
      name              = data_disks.value.name != "" ? data_disks.value.name : null
      performance_level = data_disks.value.performance_level != "" ? data_disks.value.performance_level : null
      encrypted         = data_disks.value.encrypted != "" ? data_disks.value.encrypted : null
      kms_key_id        = data_disks.value.kms_key_id != "" ? data_disks.value.kms_key_id : null
    }
  }

  security_group_ids         = length(var.spec.security_group_ids) > 0 ? var.spec.security_group_ids : null
  internet_max_bandwidth_out = var.spec.internet_max_bandwidth_out > 0 ? var.spec.internet_max_bandwidth_out : null
  internet_charge_type       = var.spec.internet_charge_type != "" ? var.spec.internet_charge_type : null

  key_name = var.spec.key_name != "" ? var.spec.key_name : null
  password = var.spec.password != "" ? var.spec.password : null

  dynamic "labels" {
    for_each = var.spec.labels
    content {
      key   = labels.key
      value = labels.value
    }
  }

  dynamic "taints" {
    for_each = var.spec.taints
    content {
      key    = taints.value.key
      value  = taints.value.value != "" ? taints.value.value : null
      effect = taints.value.effect != "" ? taints.value.effect : null
    }
  }

  cpu_policy            = var.spec.cpu_policy != "" ? var.spec.cpu_policy : null
  runtime_name          = var.spec.runtime_name != "" ? var.spec.runtime_name : null
  runtime_version       = var.spec.runtime_version != "" ? var.spec.runtime_version : null
  unschedulable         = var.spec.unschedulable
  user_data             = var.spec.user_data != "" ? var.spec.user_data : null
  install_cloud_monitor = var.spec.install_cloud_monitor

  dynamic "scaling_config" {
    for_each = var.spec.scaling_config != null ? [var.spec.scaling_config] : []
    content {
      enable   = scaling_config.value.enable
      min_size = scaling_config.value.min_size
      max_size = scaling_config.value.max_size
      type     = scaling_config.value.type != "" ? scaling_config.value.type : null
    }
  }

  multi_az_policy = var.spec.multi_az_policy != "" ? var.spec.multi_az_policy : null

  dynamic "management" {
    for_each = var.spec.management != null ? [var.spec.management] : []
    content {
      enable          = management.value.enable
      auto_repair     = management.value.auto_repair
      auto_upgrade    = management.value.auto_upgrade
      max_unavailable = management.value.max_unavailable
    }
  }

  spot_strategy = var.spec.spot_strategy != "" ? var.spec.spot_strategy : null

  dynamic "spot_price_limit" {
    for_each = var.spec.spot_price_limits
    content {
      instance_type = spot_price_limit.value.instance_type
      price_limit   = spot_price_limit.value.price_limit
    }
  }

  instance_charge_type = var.spec.instance_charge_type
  period               = var.spec.period
  auto_renew           = var.spec.auto_renew
  auto_renew_period    = var.spec.auto_renew_period

  resource_group_id = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  ram_role_name     = var.spec.ram_role_name != "" ? var.spec.ram_role_name : null
}
