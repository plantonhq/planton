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
  description = "OpenStackRouterInterfaceSpec defines the configuration for attaching a router to a subnet"
  type = object({
    # (Required) The ID of the router to attach the subnet to.
    # Supports StringValueOrRef pattern - use {value: "router-id"} for literal values.
    router_id = object({
      value = string
    })

    # (Required) The ID of the subnet to connect to the router.
    # Supports StringValueOrRef pattern - use {value: "subnet-id"} for literal values.
    subnet_id = object({
      value = string
    })

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
