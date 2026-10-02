locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The collection's GCP id and display name default to metadata.name (the
  # spec-level contract) -- identical to the Pulumi module.
  collection_id           = var.spec.collection_id != "" ? var.spec.collection_id : var.metadata.name
  collection_display_name = var.spec.collection_display_name != "" ? var.spec.collection_display_name : var.metadata.name

  # Empty optional strings, lists, and maps become null so the provider
  # omits them from the API payload instead of sending empty values it
  # would reject or diff on. Exactly one of params / json_params is set (a
  # spec rule).
  params                       = length(var.spec.params) > 0 ? var.spec.params : null
  json_params                  = var.spec.json_params != "" ? var.spec.json_params : null
  incremental_refresh_interval = var.spec.incremental_refresh_interval != "" ? var.spec.incremental_refresh_interval : null
  sync_mode                    = var.spec.sync_mode != "" ? var.spec.sync_mode : null
  connector_modes              = length(var.spec.connector_modes) > 0 ? var.spec.connector_modes : null
  kms_key_name                 = var.spec.kms_key_name != "" ? var.spec.kms_key_name : null
  deletion_policy              = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
