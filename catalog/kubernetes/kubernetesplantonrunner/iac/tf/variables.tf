variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "KubernetesPlantonRunner specification"
  type = object({
    namespace              = string
    create_namespace       = optional(bool, false)
    token                  = string
    runner_name            = optional(string, "")
    control_plane_endpoint = optional(string, "")
    runner_version         = optional(string)
    image_repository       = optional(string)
    chart_version          = optional(string, "")
    chart_repository       = optional(string, "")
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
    build = optional(object({
      enabled          = optional(bool, false)
      tekton_namespace = optional(string, "")
    }))
    helm_values = optional(string, "")
  })
}