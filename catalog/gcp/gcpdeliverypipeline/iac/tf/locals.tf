locals {
  # Honor the spec contract: an empty project_id falls back to the
  # provider's default project. Passing null (instead of "") lets the
  # google provider resolve its own project from configuration.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The pipeline ID defaults to metadata.name -- identical to the Pulumi
  # module.
  delivery_pipeline_id = var.spec.delivery_pipeline_id != "" ? var.spec.delivery_pipeline_id : var.metadata.name
  deletion_policy      = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The same planton-ai_* label set the Pulumi module applies, on the
  # pipeline and on every automation. User labels merge in first so the
  # platform attribution labels can never be clobbered by a spec label with
  # the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpdeliverypipeline"
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

  attribution_labels = merge(local.base_labels, local.org_label, local.env_label, local.id_label)
  final_labels       = merge(var.spec.labels, local.attribution_labels)

  # The stages, empty when no serial pipeline is declared.
  stages = var.spec.serial_pipeline != null ? var.spec.serial_pipeline.stages : []

  # The automations, keyed by their own declared IDs so adding or removing
  # one never renames or recreates its siblings.
  automations = { for automation in var.spec.automations : automation.automation_id => automation }
}
