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
  description = "OpenStackServerGroupSpec defines the configuration for a compute server group"
  type = object({
    # (Required) Placement policy: "affinity", "anti-affinity", "soft-affinity", "soft-anti-affinity"
    policy = string

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
