# OpenFgaAuthorizationModel Main Resources
# This file creates the OpenFGA authorization model resource.
#
# An authorization model defines the types, relations, and access rules for
# fine-grained authorization within an OpenFGA store.
#
# Note: Authorization models are immutable. Changing the model will
# create a new model (new ID) rather than updating the existing one.
#
# Reference: https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/authorization_model

# Convert DSL to JSON if model_dsl is provided.
# The openfga_authorization_model_document data source handles the conversion
# and produces a stable JSON output regardless of input format.
data "openfga_authorization_model_document" "dsl_to_json" {
  count = local.use_dsl ? 1 : 0
  dsl   = local.model_dsl
}

# Create the authorization model resource.
# Uses either the converted JSON from DSL or the directly provided JSON.
resource "openfga_authorization_model" "this" {
  store_id   = local.store_id
  model_json = local.model_json_final
}
