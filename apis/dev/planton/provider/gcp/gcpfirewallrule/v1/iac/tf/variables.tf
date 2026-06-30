variable "spec" {
  description = "GcpFirewallRule spec"
  type = object({
    project_id = object({
      value = string
    })
    network = object({
      value = string
    })
    rule_name = string
    direction = string
    action    = string
    rules = list(object({
      protocol = string
      ports    = optional(list(string), [])
    }))
    priority                = optional(number, 1000)
    description             = optional(string, "")
    source_ranges           = optional(list(string), [])
    destination_ranges      = optional(list(string), [])
    source_tags             = optional(list(string), [])
    target_tags             = optional(list(string), [])
    source_service_accounts = optional(list(string), [])
    target_service_accounts = optional(list(string), [])
    disabled                = optional(bool, false)
    log_config = optional(object({
      metadata = string
    }), null)
  })
}

variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key = optional(string, "")
  })
  default = {
    service_account_key = ""
  }
}
