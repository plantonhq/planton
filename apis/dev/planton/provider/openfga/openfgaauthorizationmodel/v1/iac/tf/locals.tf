# OpenFgaAuthorizationModel Local Values
# This file computes local values from the input variables.

locals {
  # Store ID extracted from the StringValueOrRef object
  store_id = var.spec.store_id.value

  # Model DSL from spec (may be null)
  model_dsl = var.spec.model_dsl

  # Model JSON from spec (may be null)
  model_json = var.spec.model_json

  # Determine if DSL is provided
  use_dsl = local.model_dsl != null && local.model_dsl != ""

  # Final model JSON to use in the resource:
  # - If DSL is provided, use the converted JSON from the data source
  # - Otherwise, use the provided JSON directly
  model_json_final = local.use_dsl ? data.openfga_authorization_model_document.dsl_to_json[0].result : local.model_json
}
