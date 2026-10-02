locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The template ID defaults to metadata.name -- the same fallback the
  # Pulumi module applies.
  template_id = var.spec.template_id != "" ? var.spec.template_id : var.metadata.name

  # Empty optional strings become null so the provider omits them.
  description      = var.spec.description != "" ? var.spec.description : null
  maximum_lifetime = var.spec.maximum_lifetime != "" ? var.spec.maximum_lifetime : null
  deletion_policy  = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The same planton-ai_* label set the Pulumi module applies, so the
  # template is attributable to its Planton object regardless of the engine
  # that created it. User labels merge in first so the platform attribution
  # labels can never be clobbered by a spec label with the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpprivatecacertificatetemplate"
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
