variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key = string
  })
}

variable "metadata" {
  description = "Resource metadata"
  type = object({
    name = string
    id   = optional(string, "")
    org  = optional(string, "")
    env  = optional(string, "")
  })
}

variable "spec" {
  description = "GcpBigtableInstance specification"
  type = object({
    project_id          = string
    instance_name       = string
    display_name        = optional(string, "")
    deletion_protection = optional(bool, true)
    force_destroy       = optional(bool, false)
    clusters = list(object({
      cluster_id         = string
      zone               = string
      num_nodes          = optional(number, 0)
      storage_type       = optional(string, "SSD")
      kms_key_name       = optional(string, "")
      node_scaling_factor = optional(string, "")
      autoscaling_config = optional(object({
        min_nodes      = number
        max_nodes      = number
        cpu_target     = number
        storage_target = optional(number, 0)
      }), null)
    }))
  })
}
