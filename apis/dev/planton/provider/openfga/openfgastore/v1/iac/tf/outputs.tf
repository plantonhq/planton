# OpenFgaStore Outputs
# This file defines the outputs from the OpenFGA store deployment.
# These values are used to populate the stack outputs protobuf message.
#
# Reference: https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/store#attributes-reference

output "id" {
  description = "The unique identifier of the OpenFGA store"
  value       = openfga_store.this.id
}

output "name" {
  description = "The name of the OpenFGA store"
  value       = openfga_store.this.name
}
