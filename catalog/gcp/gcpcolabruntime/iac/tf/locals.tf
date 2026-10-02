locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The runtime id and display name default to metadata.name -- identical to
  # the Pulumi module. The id is always sent: the provider never reads it
  # back, so an unsent id would leave the resource without an identity.
  runtime_id   = var.spec.runtime_id != "" ? var.spec.runtime_id : var.metadata.name
  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  description     = var.spec.description != "" ? var.spec.description : null
  desired_state   = var.spec.desired_state != "" ? var.spec.desired_state : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
