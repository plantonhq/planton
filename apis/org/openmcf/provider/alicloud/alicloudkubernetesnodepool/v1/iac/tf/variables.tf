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
  description = "Alibaba Cloud ACK Kubernetes node pool specification"
  type = object({
    region         = string
    cluster_id     = string
    name           = string
    vswitch_ids    = list(string)
    instance_types = list(string)
    desired_size   = optional(number)

    image_type = optional(string, "AliyunLinux3")

    system_disk = optional(object({
      category          = optional(string, "cloud_essd")
      size              = optional(number, 120)
      performance_level = optional(string, "")
      encrypted         = optional(bool, false)
      kms_key_id        = optional(string, "")
    }))

    data_disks = optional(list(object({
      category          = optional(string, "cloud_essd")
      size              = number
      name              = optional(string, "")
      performance_level = optional(string, "")
      encrypted         = optional(string, "")
      kms_key_id        = optional(string, "")
    })), [])

    security_group_ids         = optional(list(string), [])
    internet_max_bandwidth_out = optional(number, 0)
    internet_charge_type       = optional(string, "")

    key_name = optional(string, "")
    password = optional(string, "")

    labels = optional(map(string), {})
    taints = optional(list(object({
      key    = string
      value  = optional(string, "")
      effect = optional(string, "")
    })), [])

    cpu_policy            = optional(string, "")
    runtime_name          = optional(string, "")
    runtime_version       = optional(string, "")
    unschedulable         = optional(bool)
    user_data             = optional(string, "")
    install_cloud_monitor = optional(bool, true)

    scaling_config = optional(object({
      enable   = optional(bool, true)
      min_size = number
      max_size = number
      type     = optional(string, "")
    }))

    multi_az_policy = optional(string, "")

    management = optional(object({
      enable          = optional(bool, true)
      auto_repair     = optional(bool)
      auto_upgrade    = optional(bool)
      max_unavailable = optional(number)
    }))

    spot_strategy = optional(string, "")
    spot_price_limits = optional(list(object({
      instance_type = string
      price_limit   = string
    })), [])

    instance_charge_type = optional(string, "PostPaid")
    period               = optional(number)
    auto_renew           = optional(bool)
    auto_renew_period    = optional(number)

    tags              = optional(map(string), {})
    resource_group_id = optional(string, "")
    ram_role_name     = optional(string, "")
  })

  validation {
    condition     = length(var.spec.vswitch_ids) >= 1 && length(var.spec.vswitch_ids) <= 5
    error_message = "vswitch_ids must contain 1 to 5 VSwitch IDs."
  }

  validation {
    condition     = length(var.spec.instance_types) >= 1
    error_message = "instance_types must contain at least one instance type."
  }

  validation {
    condition = contains([
      "AliyunLinux", "AliyunLinux3", "AliyunLinux3Arm64", "AliyunLinuxUEFI",
      "CentOS", "Windows", "WindowsCore", "ContainerOS", "Ubuntu",
      "AliyunLinux3ContainerOptimized", "Custom"
    ], var.spec.image_type)
    error_message = "image_type must be a valid ACK node image type."
  }

  validation {
    condition     = contains(["PostPaid", "PrePaid"], var.spec.instance_charge_type)
    error_message = "instance_charge_type must be one of: PostPaid, PrePaid."
  }
}
