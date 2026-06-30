##############################################
# variables.tf
#
# Input variables for the KubernetesGatewayApiCrds
# Terraform module.
#
# These variables mirror the spec.proto fields
# for the KubernetesGatewayApiCrds resource.
##############################################

variable "metadata" {
  description = "Resource metadata including name"
  type = object({
    name = string
  })
}

variable "spec" {
  description = "KubernetesGatewayApiCrds specification"
  type = object({
    # Gateway API version to install (e.g., "v1.2.1", "v1.3.0")
    version = optional(string, "v1.2.1")

    # Installation channel configuration
    install_channel = optional(object({
      # Channel: "standard" or "experimental"
      channel = optional(string, "standard")
    }), { channel = "standard" })
  })
  default = {
    version         = "v1.2.1"
    install_channel = { channel = "standard" }
  }
}
