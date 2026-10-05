variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "KubernetesOtelCollector specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    mode             = optional(string, "")
    config_yaml      = string
    replicas         = optional(number)
    autoscaler = optional(object({
      min_replicas              = optional(number)
      max_replicas              = number
      target_cpu_utilization    = optional(number)
      target_memory_utilization = optional(number)
    }))
    image            = optional(string, "")
    service_account  = optional(string, "")
    env              = optional(map(string), {})
    env_from_secrets = optional(list(string), [])
    volumes = optional(list(object({
      name       = string
      mount_path = string
      read_only  = optional(bool, false)
      sub_path   = optional(string, "")
      config_map = optional(object({
        name         = string
        key          = optional(string, "")
        path         = optional(string, "")
        default_mode = optional(number)
      }))
      secret = optional(object({
        name         = string
        key          = optional(string, "")
        path         = optional(string, "")
        default_mode = optional(number)
      }))
      host_path = optional(object({
        path = string
        type = optional(string, "")
      }))
      empty_dir = optional(object({
        medium     = optional(string, "")
        size_limit = optional(string, "")
      }))
      pvc = optional(object({
        claim_name = string
        read_only  = optional(bool, false)
      }))
      service_account_token = optional(object({
        audience           = string
        expiration_seconds = optional(number, 0)
        path               = optional(string, "")
      }))
    })), [])
    additional_ports = optional(list(object({
      name     = string
      port     = number
      protocol = optional(string)
    })), [])
    resources = optional(object({
      limits = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
      requests = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
    }))
    scheduling = optional(object({
      node_selector = optional(map(string), {})
      tolerations = optional(list(object({
        key                = optional(string, "")
        operator           = optional(string, "")
        value              = optional(string, "")
        effect             = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
      priority_class_name = optional(string, "")
    }))
    pod_security_context = optional(object({
      run_as_user            = optional(number)
      run_as_group           = optional(number)
      run_as_non_root        = optional(bool)
      fs_group               = optional(number)
      fs_group_change_policy = optional(string, "")
      supplemental_groups    = optional(list(number), [])
      sysctls = optional(list(object({
        name  = string
        value = string
      })), [])
      seccomp_profile = optional(object({
        type              = optional(string, "")
        localhost_profile = optional(string, "")
      }))
    }))
    service_monitor_enabled = optional(bool, false)
  })
}
