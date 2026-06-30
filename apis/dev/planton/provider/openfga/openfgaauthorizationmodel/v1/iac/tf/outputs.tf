# OpenFgaAuthorizationModel Outputs
# This file defines the outputs from the OpenFGA authorization model deployment.
# These values are used to populate the stack outputs protobuf message.
#
# Reference: https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/authorization_model#attributes-reference

output "id" {
  description = "The unique identifier of the authorization model"
  value       = openfga_authorization_model.this.id
}
