locals {
  # An empty project_id falls back to the provider's default project.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # issuance_config_name falls back to metadata.name -- explicit conditional, so
  # both engines derive the identical cloud-side name.
  issuance_config_name = var.spec.issuance_config_name != "" ? var.spec.issuance_config_name : var.metadata.name

  # Empty defers to the provider default ("global"); the output reports the
  # effective location either way.
  location = var.spec.location != "" ? var.spec.location : null

  # The pool's project, location, and id, parsed from its full name
  # (projects/{project}/locations/{location}/caPools/{pool}, which the spec
  # validates) -- the agent's grant lands on the pool where it lives, whatever
  # project this config is in. Identical parse to the Pulumi module.
  ca_pool_parts    = split("/", var.spec.ca_pool)
  ca_pool_project  = local.ca_pool_parts[1]
  ca_pool_location = local.ca_pool_parts[3]
  ca_pool_id       = local.ca_pool_parts[5]

  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = local.issuance_config_name
    "planton-ai_kind"     = "gcpcertmanagerissuanceconfig"
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

  # User labels first so platform attribution labels win on key conflicts --
  # identical merge order to the Pulumi module.
  final_labels = merge(
    var.spec.labels,
    local.base_labels,
    local.org_label,
    local.env_label,
    local.id_label,
  )
}
