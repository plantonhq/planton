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
  description = "Alibaba Cloud ECS Instance specification"
  type = object({
    region             = string
    vswitch_id         = string
    security_group_ids = list(string)
    instance_type      = string
    image_id           = string
    instance_name      = optional(string, "")
    host_name          = optional(string, "")
    description        = optional(string, "")

    system_disk = optional(object({
      category          = optional(string, "cloud_essd")
      size              = optional(number, 40)
      performance_level = optional(string, "")
      encrypted         = optional(bool, false)
      kms_key_id        = optional(string, "")
    }), {
      category          = "cloud_essd"
      size              = 40
      performance_level = ""
      encrypted         = false
      kms_key_id        = ""
    })

    data_disks = optional(list(object({
      size                 = number
      category             = optional(string, "cloud_essd")
      name                 = optional(string, "")
      performance_level    = optional(string, "")
      encrypted            = optional(bool, false)
      kms_key_id           = optional(string, "")
      snapshot_id          = optional(string, "")
      delete_with_instance = optional(bool, true)
      description          = optional(string, "")
    })), [])

    key_name = optional(string, "")
    password = optional(string, "")

    internet_max_bandwidth_out = optional(number, 0)
    internet_charge_type       = optional(string, "")
    instance_charge_type       = optional(string, "PostPaid")
    period                     = optional(number, null)
    period_unit                = optional(string, "")

    spot_strategy    = optional(string, "")
    spot_price_limit = optional(number, null)

    user_data                       = optional(string, "")
    role_name                       = optional(string, "")
    deletion_protection             = optional(bool, false)
    security_enhancement_strategy   = optional(string, "")
    resource_group_id               = optional(string, "")
    tags                            = optional(map(string), {})
  })

  validation {
    condition     = can(regex("^ecs\\.", var.spec.instance_type))
    error_message = "instance_type must start with 'ecs.'"
  }

  validation {
    condition     = length(var.spec.security_group_ids) >= 1
    error_message = "At least one security group ID is required."
  }

  validation {
    condition     = contains(["PostPaid", "PrePaid"], var.spec.instance_charge_type)
    error_message = "instance_charge_type must be one of: PostPaid, PrePaid."
  }

  validation {
    condition = (
      var.spec.internet_charge_type == "" ||
      contains(["PayByTraffic", "PayByBandwidth"], var.spec.internet_charge_type)
    )
    error_message = "internet_charge_type must be one of: PayByTraffic, PayByBandwidth."
  }

  validation {
    condition = (
      var.spec.spot_strategy == "" ||
      contains(["NoSpot", "SpotAsPriceGo", "SpotWithPriceLimit"], var.spec.spot_strategy)
    )
    error_message = "spot_strategy must be one of: NoSpot, SpotAsPriceGo, SpotWithPriceLimit."
  }

  validation {
    condition = (
      var.spec.security_enhancement_strategy == "" ||
      contains(["Active", "Deactive"], var.spec.security_enhancement_strategy)
    )
    error_message = "security_enhancement_strategy must be one of: Active, Deactive."
  }

  validation {
    condition = (
      var.spec.system_disk.category == "" ||
      contains(["cloud_efficiency", "cloud_ssd", "cloud_essd", "cloud_auto", "cloud_essd_entry"], var.spec.system_disk.category)
    )
    error_message = "system_disk.category must be one of: cloud_efficiency, cloud_ssd, cloud_essd, cloud_auto, cloud_essd_entry."
  }

  validation {
    condition = (
      var.spec.period_unit == "" ||
      contains(["Week", "Month"], var.spec.period_unit)
    )
    error_message = "period_unit must be one of: Week, Month."
  }
}
