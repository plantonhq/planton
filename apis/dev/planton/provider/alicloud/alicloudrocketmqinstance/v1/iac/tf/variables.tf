variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name   = string
    id     = optional(string)
    org    = optional(string)
    env    = optional(string)
    labels = optional(map(string))
    tags   = optional(list(string))
  })
}

variable "spec" {
  description = "Alibaba Cloud RocketMQ 5.x Instance specification"
  type = object({
    region          = string
    series_code     = string
    sub_series_code = string
    vpc_id          = string
    instance_name   = optional(string, "")
    remark          = optional(string, "")

    payment_type     = optional(string, "PayAsYouGo")
    period           = optional(number, null)
    period_unit      = optional(string, "")
    auto_renew       = optional(bool, null)
    auto_renew_period = optional(number, null)

    vswitch_id        = optional(string, "")
    security_group_id = optional(string, "")
    internet_info = optional(object({
      enabled            = optional(bool, false)
      flow_out_type      = optional(string, "payByTraffic")
      flow_out_bandwidth = optional(number, null)
    }), null)

    msg_process_spec = optional(string, "")
    product_info = optional(object({
      message_retention_time = optional(number, null)
      send_receive_ratio     = optional(number, null)
      auto_scaling           = optional(bool, null)
      trace_on               = optional(bool, null)
      storage_encryption     = optional(bool, null)
      storage_secret_key     = optional(string, "")
    }), null)

    ip_whitelists     = optional(list(string), [])
    resource_group_id = optional(string, "")
    tags              = optional(map(string), {})

    topics = optional(list(object({
      topic_name   = string
      message_type = optional(string, "NORMAL")
      remark       = optional(string, "")
      max_send_tps = optional(number, null)
    })), [])

    consumer_groups = optional(list(object({
      consumer_group_id  = string
      delivery_order_type = optional(string, "")
      remark             = optional(string, "")
      max_receive_tps    = optional(number, null)
      consume_retry_policy = optional(object({
        retry_policy            = optional(string, "DefaultRetryPolicy")
        max_retry_times         = optional(number, null)
        dead_letter_target_topic = optional(string, "")
      }), null)
    })), [])
  })

  validation {
    condition     = contains(["standard", "professional", "ultimate"], var.spec.series_code)
    error_message = "series_code must be one of: standard, professional, ultimate."
  }

  validation {
    condition     = contains(["cluster_ha", "single_node", "serverless"], var.spec.sub_series_code)
    error_message = "sub_series_code must be one of: cluster_ha, single_node, serverless."
  }

  validation {
    condition     = contains(["PayAsYouGo", "Subscription"], var.spec.payment_type)
    error_message = "payment_type must be one of: PayAsYouGo, Subscription."
  }
}
