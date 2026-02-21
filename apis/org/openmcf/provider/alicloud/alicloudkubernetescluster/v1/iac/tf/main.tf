resource "alicloud_cs_managed_kubernetes" "cluster" {
  name               = local.cluster_name
  cluster_spec       = var.spec.cluster_spec
  vswitch_ids        = var.spec.vswitch_ids
  service_cidr       = var.spec.service_cidr
  proxy_mode         = var.spec.proxy_mode
  node_cidr_mask     = var.spec.node_cidr_mask
  new_nat_gateway    = var.spec.new_nat_gateway
  slb_internet_enabled = var.spec.slb_internet_enabled
  enable_rrsa        = var.spec.enable_rrsa
  deletion_protection = var.spec.deletion_protection
  tags               = local.final_tags

  version            = var.spec.version != "" ? var.spec.version : null
  cluster_domain     = var.spec.cluster_domain != "" ? var.spec.cluster_domain : null
  pod_cidr           = var.spec.pod_cidr != "" ? var.spec.pod_cidr : null
  pod_vswitch_ids    = length(var.spec.pod_vswitch_ids) > 0 ? var.spec.pod_vswitch_ids : null
  security_group_id  = var.spec.security_group_id != "" ? var.spec.security_group_id : null
  is_enterprise_security_group = var.spec.is_enterprise_security_group
  encryption_provider_key = var.spec.encryption_provider_key != "" ? var.spec.encryption_provider_key : null
  custom_san         = var.spec.custom_san != "" ? var.spec.custom_san : null
  resource_group_id  = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  timezone           = var.spec.timezone != "" ? var.spec.timezone : null

  # Control plane logging
  control_plane_log_project    = try(var.spec.logging.control_plane_log_project, "") != "" ? var.spec.logging.control_plane_log_project : null
  control_plane_log_ttl        = try(var.spec.logging.control_plane_log_ttl, "30")
  control_plane_log_components = try(length(var.spec.logging.control_plane_log_components), 0) > 0 ? var.spec.logging.control_plane_log_components : null

  dynamic "audit_log_config" {
    for_each = try(var.spec.logging.audit_log_enabled, false) ? [1] : []
    content {
      enabled          = true
      sls_project_name = try(var.spec.logging.audit_log_sls_project, "") != "" ? var.spec.logging.audit_log_sls_project : null
    }
  }

  dynamic "addons" {
    for_each = local.addons_map
    content {
      name     = addons.value.name
      config   = addons.value.config != "" ? addons.value.config : null
      version  = addons.value.version != "" ? addons.value.version : null
      disabled = addons.value.disabled
    }
  }

  dynamic "maintenance_window" {
    for_each = var.spec.maintenance_window != null ? [var.spec.maintenance_window] : []
    content {
      enable           = maintenance_window.value.enable
      maintenance_time = maintenance_window.value.maintenance_time
      duration         = maintenance_window.value.duration
      weekly_period    = maintenance_window.value.weekly_period
    }
  }

  dynamic "operation_policy" {
    for_each = var.spec.auto_upgrade != null && var.spec.auto_upgrade.enabled ? [var.spec.auto_upgrade] : []
    content {
      cluster_auto_upgrade {
        enabled = true
        channel = operation_policy.value.channel
      }
    }
  }
}
