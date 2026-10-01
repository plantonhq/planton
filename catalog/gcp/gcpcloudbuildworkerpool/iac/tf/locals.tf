locals {
  # Honor the spec contract: an empty project_id falls back to the
  # provider's default project. Passing null (instead of "") lets the
  # google provider resolve its own project from configuration.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The pool ID defaults to metadata.name -- identical to the Pulumi module.
  worker_pool_id  = var.spec.worker_pool_id != "" ? var.spec.worker_pool_id : var.metadata.name
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Google wants the peered network as projects/{NUMBER}/global/networks/{name}.
  # A GcpVpcNetwork reference arrives as a self-link carrying the project
  # ID: the prefix is stripped and, when the project segment is not already
  # a number, the number is resolved through one guarded project lookup
  # (main.tf) -- the same rule as the Pulumi module.
  peered_network_value        = var.spec.network_config != null ? var.spec.network_config.peered_network : ""
  peered_network_path         = trimprefix(local.peered_network_value, "https://www.googleapis.com/compute/v1/")
  peered_network_segments     = split("/", local.peered_network_path)
  peered_network_project      = local.peered_network_value != "" ? local.peered_network_segments[1] : ""
  peered_network_name         = local.peered_network_value != "" ? local.peered_network_segments[length(local.peered_network_segments) - 1] : ""
  peered_network_needs_number = local.peered_network_project != "" && !can(regex("^[0-9]+$", local.peered_network_project))
  peered_network = (
    local.peered_network_value == "" ? null :
    local.peered_network_needs_number ? "projects/${data.google_project.peered_network[0].number}/global/networks/${local.peered_network_name}" :
    local.peered_network_path
  )
}
