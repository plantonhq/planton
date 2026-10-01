locals {
  # An empty project_id falls back to the provider's default project.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # trust_config_name falls back to metadata.name -- explicit conditional, so
  # both engines derive the identical cloud-side name.
  trust_config_name = var.spec.trust_config_name != "" ? var.spec.trust_config_name : var.metadata.name

  # The provider requires a location; "global" serves global load balancers.
  location = var.spec.location != "" ? var.spec.location : "global"

  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = local.trust_config_name
    "planton-ai_kind"     = "gcpcertmanagertrustconfig"
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
