variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsLbTargetGroup specification"
  type = object({
    region = string
    vpc_id = optional(string, "")
    target_type = optional(string, "")
    port = optional(number, 0)
    protocol = optional(string, "")
    protocol_version = optional(string, "")
    ip_address_type = optional(string, "")
    health_check = optional(object({
      enabled = optional(bool)
      protocol = optional(string, "")
      port = optional(string, "")
      path = optional(string, "")
      healthy_threshold = optional(number, 0)
      unhealthy_threshold = optional(number, 0)
      interval_seconds = optional(number, 0)
      timeout_seconds = optional(number, 0)
      matcher = optional(string, "")
    }))
    stickiness = optional(object({
      type = string
      enabled = optional(bool)
      cookie_duration_seconds = optional(number, 0)
      cookie_name = optional(string, "")
    }))
    deregistration_delay_seconds = optional(number, 0)
    slow_start_seconds = optional(number, 0)
    load_balancing_algorithm_type = optional(string, "")
    load_balancing_anomaly_mitigation = optional(string, "")
    load_balancing_cross_zone_enabled = optional(string, "")
    preserve_client_ip = optional(bool)
    proxy_protocol_v2 = optional(bool, false)
    connection_termination = optional(bool, false)
    lambda_multi_value_headers_enabled = optional(bool, false)
    target_group_health = optional(object({
      dns_failover = optional(object({
        minimum_healthy_targets_count = optional(string, "")
        minimum_healthy_targets_percentage = optional(string, "")
      }))
      unhealthy_state_routing = optional(object({
        minimum_healthy_targets_count = optional(number, 0)
        minimum_healthy_targets_percentage = optional(string, "")
      }))
    }))
    target_health_state = optional(object({
      enable_unhealthy_connection_termination = optional(bool, false)
      unhealthy_draining_interval_seconds = optional(number, 0)
    }))
    targets = optional(list(object({
      target_id = string
      port = optional(number, 0)
      availability_zone = optional(string, "")
      quic_server_id = optional(string, "")
    })), [])
    target_control_port = optional(number, 0)
  })
}
