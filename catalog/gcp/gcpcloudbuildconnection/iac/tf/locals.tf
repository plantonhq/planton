locals {
  # Honor the spec contract: an empty project_id falls back to the
  # provider's default project. Passing null (instead of "") lets the
  # google provider resolve its own project from configuration.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The connection ID defaults to metadata.name -- identical to the Pulumi
  # module.
  connection_id   = var.spec.connection_id != "" ? var.spec.connection_id : var.metadata.name
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
