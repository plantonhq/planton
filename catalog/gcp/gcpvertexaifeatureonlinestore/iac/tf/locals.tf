locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Empty optional strings become null so the provider omits them from the
  # API payload instead of sending empty values it would reject or diff on.
  kms_key_name    = var.spec.kms_key_name != "" ? var.spec.kms_key_name : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The same planton-ai_* label set the Pulumi module applies, on the store
  # and on every feature view. User labels merge in first so the platform
  # attribution labels can never be clobbered by a spec label with the same
  # key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.spec.feature_online_store_id
    "planton-ai_kind"     = "gcpvertexaifeatureonlinestore"
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

  # The folded feature views keyed by feature_view_id, so adding or removing
  # one never renumbers the others.
  feature_views = { for view in var.spec.feature_views : view.feature_view_id => view }
}
