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
  description = "Alibaba Cloud ACK Managed Kubernetes cluster specification"
  type = object({
    region         = string
    name           = optional(string, "")
    version        = optional(string, "")
    cluster_spec   = optional(string, "ack.standard")
    cluster_domain = optional(string, "")

    vswitch_ids      = list(string)
    pod_cidr         = optional(string, "")
    pod_vswitch_ids  = optional(list(string), [])
    service_cidr     = string
    proxy_mode       = optional(string, "ipvs")
    node_cidr_mask   = optional(number, 24)
    new_nat_gateway  = optional(bool, true)
    slb_internet_enabled = optional(bool, true)

    security_group_id            = optional(string, "")
    is_enterprise_security_group = optional(bool, false)
    enable_rrsa                  = optional(bool, false)
    deletion_protection          = optional(bool, false)
    encryption_provider_key      = optional(string, "")
    custom_san                   = optional(string, "")

    addons = optional(list(object({
      name     = string
      config   = optional(string, "")
      version  = optional(string, "")
      disabled = optional(bool, false)
    })), [])

    logging = optional(object({
      control_plane_log_project    = optional(string, "")
      control_plane_log_ttl        = optional(string, "30")
      control_plane_log_components = optional(list(string), [])
      audit_log_enabled            = optional(bool, false)
      audit_log_sls_project        = optional(string, "")
    }))

    maintenance_window = optional(object({
      enable           = bool
      maintenance_time = string
      duration         = string
      weekly_period    = optional(string, "Thursday")
    }))

    auto_upgrade = optional(object({
      enabled = bool
      channel = optional(string, "patch")
    }))

    tags              = optional(map(string), {})
    resource_group_id = optional(string, "")
    timezone          = optional(string, "")
  })

  validation {
    condition     = length(var.spec.vswitch_ids) >= 1 && length(var.spec.vswitch_ids) <= 5
    error_message = "vswitch_ids must contain 1 to 5 VSwitch IDs."
  }

  validation {
    condition     = contains(["ack.standard", "ack.pro.small"], var.spec.cluster_spec)
    error_message = "cluster_spec must be one of: ack.standard, ack.pro.small."
  }

  validation {
    condition     = contains(["iptables", "ipvs"], var.spec.proxy_mode)
    error_message = "proxy_mode must be one of: iptables, ipvs."
  }

  validation {
    condition     = var.spec.node_cidr_mask >= 24 && var.spec.node_cidr_mask <= 28
    error_message = "node_cidr_mask must be between 24 and 28."
  }
}
