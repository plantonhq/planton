locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The request's id defaults to metadata.name -- identical to the Pulumi
  # module.
  queued_resource_id = var.spec.queued_resource_id != "" ? var.spec.queued_resource_id : var.metadata.name

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Google wants each node's parent as projects/{project}/locations/{zone}.
  # The project is the spec's, or the provider's own when the spec names
  # none (read from provider configuration, no API call) -- the same rule
  # as the Pulumi module.
  node_project = local.project_id != null ? local.project_id : data.google_client_config.current[0].project
  node_parent  = "projects/${local.node_project}/locations/${var.spec.zone}"

  # A GcpSubnetwork reference arrives as a compute self-link; Cloud TPU takes
  # the relative path.
  compute_prefix = "https://www.googleapis.com/compute/v1/"
}
