# OpenFgaAuthorizationModel Variables
# This file defines all input variables for the OpenFgaAuthorizationModel Terraform module.
# These variables map to the OpenFgaAuthorizationModelSpec protobuf message.

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
  description = "OpenFgaAuthorizationModel specification"
  type = object({
    # store_id is the unique identifier of the OpenFGA store where this model will be created.
    # This is an object with a 'value' field containing the store ID.
    # The value is resolved by Planton runtime if using foreign key references.
    # Note: Changing the store_id requires replacing the model.
    store_id = object({
      value = string
    })

    # model_dsl is the authorization model definition in DSL format (recommended).
    # The DSL format is more human-readable than JSON.
    # Exactly one of model_dsl or model_json must be specified.
    model_dsl = optional(string)

    # model_json is the authorization model definition in JSON format.
    # Use this if you prefer JSON over DSL, or if you're migrating from existing JSON models.
    # Exactly one of model_dsl or model_json must be specified.
    model_json = optional(string)
  })

  validation {
    condition = (
      (var.spec.model_dsl != null && var.spec.model_dsl != "" && (var.spec.model_json == null || var.spec.model_json == "")) ||
      (var.spec.model_json != null && var.spec.model_json != "" && (var.spec.model_dsl == null || var.spec.model_dsl == ""))
    )
    error_message = "Exactly one of model_dsl or model_json must be specified."
  }
}
