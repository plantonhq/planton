resource "alicloud_rocketmq_instance" "main" {
  instance_name  = local.instance_name
  series_code    = var.spec.series_code
  sub_series_code = var.spec.sub_series_code
  service_code   = "rmq"
  payment_type   = var.spec.payment_type
  commodity_code = local.commodity_code
  remark         = var.spec.remark != "" ? var.spec.remark : null
  period         = var.spec.period
  period_unit    = var.spec.period_unit != "" ? var.spec.period_unit : null
  auto_renew     = var.spec.auto_renew
  auto_renew_period = var.spec.auto_renew_period
  ip_whitelists  = length(var.spec.ip_whitelists) > 0 ? var.spec.ip_whitelists : null
  resource_group_id = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags           = local.final_tags

  network_info {
    vpc_info {
      vpc_id = var.spec.vpc_id

      dynamic "vswitches" {
        for_each = var.spec.vswitch_id != "" ? [var.spec.vswitch_id] : []
        content {
          vswitch_id = vswitches.value
        }
      }

      security_group_ids = var.spec.security_group_id != "" ? var.spec.security_group_id : null
    }

    internet_info {
      internet_spec      = local.internet_spec
      flow_out_type      = local.flow_out_type
      flow_out_bandwidth = (
        local.internet_enabled && var.spec.internet_info != null
        ? var.spec.internet_info.flow_out_bandwidth
        : null
      )
    }
  }

  dynamic "product_info" {
    for_each = var.spec.msg_process_spec != "" || var.spec.product_info != null ? [1] : []
    content {
      msg_process_spec     = var.spec.msg_process_spec
      message_retention_time = var.spec.product_info != null ? var.spec.product_info.message_retention_time : null
      send_receive_ratio   = var.spec.product_info != null ? var.spec.product_info.send_receive_ratio : null
      auto_scaling         = var.spec.product_info != null ? var.spec.product_info.auto_scaling : null
      trace_on             = var.spec.product_info != null ? var.spec.product_info.trace_on : null
      storage_encryption   = var.spec.product_info != null ? var.spec.product_info.storage_encryption : null
      storage_secret_key   = (
        var.spec.product_info != null && var.spec.product_info.storage_secret_key != ""
        ? var.spec.product_info.storage_secret_key
        : null
      )
    }
  }
}
