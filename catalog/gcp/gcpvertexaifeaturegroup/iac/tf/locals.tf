locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Empty optional strings become null so the provider omits them from the
  # API payload instead of sending empty values it would reject or diff on.
  description     = var.spec.description != "" ? var.spec.description : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # A GcpBigQueryTable reference resolves to project.dataset.table; Google
  # stores the bq:// form, so the prefix is added when missing -- identical
  # to the Pulumi module.
  input_uri = (
    var.spec.big_query == null ? null :
    startswith(var.spec.big_query.input_uri, "bq://") ? var.spec.big_query.input_uri : "bq://${var.spec.big_query.input_uri}"
  )

  # The same planton-ai_* label set the Pulumi module applies, on the group
  # and on every registered feature. User labels merge in first so the
  # platform attribution labels can never be clobbered by a spec label with
  # the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.spec.feature_group_id
    "planton-ai_kind"     = "gcpvertexaifeaturegroup"
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

  # The folded features keyed by feature_id, so adding or removing one never
  # renumbers the others.
  features = { for feature in var.spec.features : feature.feature_id => feature }
}
