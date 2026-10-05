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
  description = "DigitalOceanLoadBalancer specification"
  type = object({
    load_balancer_name = string
    region = optional(string, "")
    vpc = optional(string, "")
    forwarding_rules = optional(list(object({
      entry_port = number
      entry_protocol = string
      target_port = number
      target_protocol = string
      tls_passthrough = optional(bool, false)
      certificate_name = optional(string, "")
    })), [])
    health_check = optional(object({
      port = number
      protocol = string
      path = optional(string, "")
      check_interval_sec = optional(number, 0)
      response_timeout_seconds = optional(number, 0)
      unhealthy_threshold = optional(number, 0)
      healthy_threshold = optional(number, 0)
    }))
    droplet_ids = optional(list(string), [])
    droplet_tag = optional(string, "")
    sticky_sessions = optional(object({
      type = string
      cookie_name = optional(string, "")
      cookie_ttl_seconds = optional(number, 0)
    }))
    type = optional(string, "")
    size = optional(string, "")
    size_unit = optional(number, 0)
    redirect_http_to_https = optional(bool, false)
    enable_proxy_protocol = optional(bool, false)
    enable_backend_keepalive = optional(bool, false)
    disable_lets_encrypt_dns_records = optional(bool, false)
    http_idle_timeout_seconds = optional(number, 0)
    tls_cipher_policy = optional(string, "")
    network = optional(string, "")
    network_stack = optional(string, "")
    project_id = optional(string, "")
    subnet_uuid = optional(string, "")
    ip = optional(string, "")
    target_load_balancer_ids = optional(list(string), [])
    firewall = optional(object({
      allow = optional(list(string), [])
      deny = optional(list(string), [])
    }))
    domains = optional(list(object({
      name = string
      is_managed = optional(bool, false)
      certificate_name = optional(string, "")
    })), [])
    glb_settings = optional(object({
      target_protocol = string
      target_port = number
      region_priorities = optional(map(number), {})
      failover_threshold = optional(number, 0)
      cdn = optional(object({
        is_enabled = optional(bool, false)
      }))
    }))
  })
}