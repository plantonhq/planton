locals {
  # Honor the spec contract: an empty project_id falls back to the
  # provider's default project. Passing null (instead of "") lets the
  # google provider resolve its own project from configuration.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The trigger name defaults to metadata.name; an empty location leaves
  # the provider's "global" -- identical to the Pulumi module.
  trigger_name    = var.spec.trigger_name != "" ? var.spec.trigger_name : var.metadata.name
  location        = var.spec.location != "" ? var.spec.location : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
