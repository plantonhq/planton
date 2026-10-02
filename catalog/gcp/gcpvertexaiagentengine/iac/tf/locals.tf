locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Google requires a display name; Planton's object name is the honest
  # default, so a manifest only sets display_name when the two must differ
  # -- identical to the Pulumi module.
  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  # Empty optional strings become null so the provider omits them from the
  # API payload instead of sending empty values it would reject or diff on.
  description     = var.spec.description != "" ? var.spec.description : null
  kms_key_name    = var.spec.kms_key_name != "" ? var.spec.kms_key_name : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The agent block and its parts, each emitted only when declared.
  agent       = var.spec.spec
  source_code = var.spec.spec != null ? var.spec.spec.source_code_spec : null
  deployment  = var.spec.spec != null ? var.spec.spec.deployment_spec : null
  memory_bank = var.spec.context_spec != null ? var.spec.context_spec.memory_bank_config : null

  # The same planton-ai_* label set the Pulumi module applies, so an agent
  # is attributable to its Planton object regardless of the engine that
  # created it. User labels merge in first so the platform attribution
  # labels can never be clobbered by a spec label with the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpvertexaiagentengine"
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
