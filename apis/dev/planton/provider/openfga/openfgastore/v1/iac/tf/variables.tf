# OpenFgaStore Variables
# This file defines all input variables for the OpenFgaStore Terraform module.
# These variables map to the OpenFgaStoreSpec protobuf message.

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
  description = "OpenFgaStore specification"
  type = object({
    # name is the display name of the OpenFGA store.
    # This name is used to identify the store in the OpenFGA server.
    # The store name is immutable - changing it requires replacing the store.
    name = string
  })
}
