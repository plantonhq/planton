locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Google requires display names on the schedule and its notebook run; both
  # default to metadata.name through the schedule's display name --
  # identical to the Pulumi module.
  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  start_time      = var.spec.start_time != "" ? var.spec.start_time : null
  end_time        = var.spec.end_time != "" ? var.spec.end_time : null
  desired_state   = var.spec.desired_state != "" ? var.spec.desired_state : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Google types the three run counts as decimal strings.
  max_concurrent_run_count        = tostring(var.spec.max_concurrent_run_count)
  max_run_count                   = var.spec.max_run_count != 0 ? tostring(var.spec.max_run_count) : null
  max_concurrent_active_run_count = var.spec.max_concurrent_active_run_count != 0 ? tostring(var.spec.max_concurrent_active_run_count) : null

  notebook = var.spec.notebook_execution_job
  pipeline = var.spec.pipeline_job

  # A GcpSubnetwork reference arrives as a compute self-link; Vertex AI takes
  # the relative path -- the same trim as the Pulumi module.
  notebook_subnetwork = (
    local.notebook != null && local.notebook.custom_environment_spec != null
    ? (local.notebook.custom_environment_spec.network_spec != null
      ? (local.notebook.custom_environment_spec.network_spec.subnetwork != ""
        ? trimprefix(local.notebook.custom_environment_spec.network_spec.subnetwork, "https://www.googleapis.com/compute/v1/")
      : null)
    : null)
    : null
  )

  # A pipeline job's peered network must read projects/{NUMBER}/global/networks/{name}.
  # A GcpVpcNetwork reference arrives as a self-link carrying the project
  # ID: the prefix is stripped and, when the project segment is not already
  # a number, the number is resolved through one guarded project lookup
  # (main.tf) -- the same rule as the Pulumi module.
  pipeline_network_raw  = local.pipeline != null ? local.pipeline.network : ""
  pipeline_network_path = trimprefix(local.pipeline_network_raw, "https://www.googleapis.com/compute/v1/")
  pipeline_network_segs = split("/", local.pipeline_network_path)
  pipeline_network_proj = local.pipeline_network_raw != "" ? local.pipeline_network_segs[1] : ""
  pipeline_network_name = local.pipeline_network_raw != "" ? local.pipeline_network_segs[length(local.pipeline_network_segs) - 1] : ""
  pipeline_network_needs_number = (
    local.pipeline_network_proj != "" && !can(regex("^[0-9]+$", local.pipeline_network_proj))
  )
  pipeline_network = (
    local.pipeline_network_raw == "" ? null :
    local.pipeline_network_needs_number ? "projects/${data.google_project.pipeline_network[0].number}/global/networks/${local.pipeline_network_name}" :
    local.pipeline_network_path
  )
}
