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
  description = "Alibaba Cloud EIP Address specification"
  type = object({
    region               = string
    address_name         = optional(string, "")
    description          = optional(string, "")
    bandwidth            = optional(number, 5)
    internet_charge_type = optional(string, "PayByTraffic")
    isp                  = optional(string, "BGP")
    resource_group_id    = optional(string, "")
    tags                 = optional(map(string), {})
  })

  validation {
    condition     = var.spec.bandwidth >= 1 && var.spec.bandwidth <= 1000
    error_message = "bandwidth must be between 1 and 1000 Mbps."
  }

  validation {
    condition     = contains(["PayByTraffic", "PayByBandwidth"], var.spec.internet_charge_type)
    error_message = "internet_charge_type must be one of: PayByTraffic, PayByBandwidth."
  }

  validation {
    condition = contains([
      "BGP", "BGP_PRO",
      "ChinaTelecom", "ChinaUnicom", "ChinaMobile",
      "ChinaTelecom_L2", "ChinaUnicom_L2", "ChinaMobile_L2",
      "BGP_FinanceCloud", "BGP_International"
    ], var.spec.isp)
    error_message = "isp must be one of: BGP, BGP_PRO, ChinaTelecom, ChinaUnicom, ChinaMobile, ChinaTelecom_L2, ChinaUnicom_L2, ChinaMobile_L2, BGP_FinanceCloud, BGP_International."
  }
}
