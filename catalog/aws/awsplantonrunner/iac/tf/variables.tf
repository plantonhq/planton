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
  description = "AwsPlantonRunner specification"
  type = object({
    region                 = string
    subnets                = list(string)
    security_groups        = optional(list(string), [])
    assign_public_ip       = optional(bool, false)
    cpu                    = optional(number)
    memory                 = optional(number)
    runner_version         = optional(string)
    image_repository       = optional(string)
    control_plane_endpoint = optional(string, "")
    token                  = string
    task_role              = optional(string, "")
    log_retention_days     = optional(number)
  })
}