locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The group's name defaults to metadata.name -- the same fallback the
  # Pulumi module applies.
  reservation_group_name = var.spec.reservation_group_name != "" ? var.spec.reservation_group_name : var.metadata.name

  location        = var.spec.location != "" ? var.spec.location : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
