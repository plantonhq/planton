locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The connection's id defaults to metadata.name -- the same fallback the
  # Pulumi module applies -- so tables and routines can name it before it
  # exists.
  connection_id = var.spec.connection_id != "" ? var.spec.connection_id : var.metadata.name

  # Empty optional strings become null so the provider omits them from the
  # API payload instead of sending values it would reject or diff on.
  location        = var.spec.location != "" ? var.spec.location : null
  friendly_name   = var.spec.friendly_name != "" ? var.spec.friendly_name : null
  description     = var.spec.description != "" ? var.spec.description : null
  kms_key_name    = var.spec.kms_key_name != "" ? var.spec.kms_key_name : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
