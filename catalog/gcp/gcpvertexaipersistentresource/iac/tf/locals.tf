locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The resource's id defaults to metadata.name (the spec-level contract) --
  # identical to the Pulumi module.
  persistent_resource_id = var.spec.persistent_resource_id != "" ? var.spec.persistent_resource_id : var.metadata.name

  # Empty optional strings become null so the provider omits them from the
  # API payload instead of sending empty values it would reject or diff on.
  display_name    = var.spec.display_name != "" ? var.spec.display_name : null
  kms_key_name    = var.spec.kms_key_name != "" ? var.spec.kms_key_name : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Google wants the peered network as projects/{NUMBER}/global/networks/{name}.
  # A GcpVpcNetwork reference arrives as a self-link carrying the project
  # ID: the prefix is stripped and, when the project segment is not already
  # a number, the number is resolved through one guarded project lookup
  # (main.tf) -- the same rule as the Pulumi module.
  network_path         = trimprefix(var.spec.network, "https://www.googleapis.com/compute/v1/")
  network_segments     = split("/", local.network_path)
  network_project      = var.spec.network != "" ? local.network_segments[1] : ""
  network_name         = var.spec.network != "" ? local.network_segments[length(local.network_segments) - 1] : ""
  network_needs_number = local.network_project != "" && !can(regex("^[0-9]+$", local.network_project))
  network = (
    var.spec.network == "" ? null :
    local.network_needs_number ? "projects/${data.google_project.network[0].number}/global/networks/${local.network_name}" :
    local.network_path
  )

  # The same planton-ai_* label set the Pulumi module applies. User labels
  # merge in first so the platform attribution labels can never be
  # clobbered by a spec label with the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = local.persistent_resource_id
    "planton-ai_kind"     = "gcpvertexaipersistentresource"
  }

  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "planton-ai_organization" = var.metadata.org } : {}

  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "planton-ai_environment" = var.metadata.env } : {}

  id_label = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "planton-ai_id" = var.metadata.id } : {}

  final_labels = merge(var.spec.labels, local.base_labels, local.org_label, local.env_label, local.id_label)
}
