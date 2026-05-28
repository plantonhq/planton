variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for KubernetesGatewayClass"
  type = object({
    # Name of the controller managing Gateways of this class (domain-prefixed path)
    controller_name = string

    # Optional reference to a controller-specific parameters resource
    parameters_ref = optional(object({
      group     = optional(string, "")
      kind      = optional(string, "")
      name      = string
      namespace = optional(string)
    }))

    # Optional human-friendly description (max 64 chars)
    description = optional(string)
  })
}
